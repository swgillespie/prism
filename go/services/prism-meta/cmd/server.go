package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	crdbpgx "github.com/cockroachdb/cockroach-go/v2/crdb/crdbpgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	metav1 "code.prism.io/proto/rpc/gen/go/prism/meta/v1"
)

var (
	port    int
	reflect bool

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run the prism-meta server",
		Run: func(cmd *cobra.Command, args []string) {
			fx.New(
				fx.Provide(zap.NewProduction),
				fx.Provide(loadConfig),
				fx.Provide(newConnectionPool),
				fx.Provide(newServer),
				fx.Provide(newGrpcServer),
				fx.Invoke(func(*grpc.Server) {}),
			).Run()
		},
	}
)

type (
	config struct {
		CockroachDBUser     string `envconfig:"COCKROACHDB_USER" default:"prism"`
		CockroachDBPassword string `envconfig:"COCKROACHDB_PASSWORD"`
		CockroachDBURL      string `envconfig:"COCKROACHDB_URL"`
		CockroachDBDatabase string `envconfig:"COCKROACHDB_DATABASE" default:"prism"`
	}

	server struct {
		metav1.UnimplementedMetaServiceServer

		pool *pgxpool.Pool
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().BoolVarP(&reflect, "reflect", "r", false, "Enable reflection")
	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
}

func (s *server) GetTableSchema(ctx context.Context, req *metav1.GetTableSchemaRequest) (*metav1.GetTableSchemaResponse, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	rows, err := conn.Query(ctx, `
		SELECT column_name, type FROM meta.table_schemas
		WHERE tenant_id = $1 AND table_name = $2
	`, req.TenantId, req.TableName)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	defer rows.Close()
	var columns []*metav1.Column
	for rows.Next() {
		var name string
		var ty int32
		if err := rows.Scan(&name, &ty); err != nil {
			return nil, status.New(codes.Internal, err.Error()).Err()
		}

		columns = append(columns, &metav1.Column{
			Name: name,
			Type: metav1.ColumnType(ty),
		})
	}

	if len(columns) == 0 {
		return nil, status.New(codes.NotFound, "table not found").Err()
	}

	return &metav1.GetTableSchemaResponse{
		TableName: req.TableName,
		Columns:   columns,
	}, nil
}

func (s *server) GetTablePartitions(ctx context.Context, req *metav1.GetTablePartitionsRequest) (*metav1.GetTablePartitionsResponse, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	var rows pgx.Rows
	if req.TimeRange != nil {
		rows, err = conn.Query(ctx,
			`SELECT partition_name FROM meta.table_partitions
                WHERE table_name = $1
                    AND (
                        start_time <= $2 AND end_time <= $3
                        OR start_time <= $2 AND end_time >= $3
                        OR start_time >= $2 AND end_time <= $3
                        OR start_time >= $2 AND end_time >= $3
                    )`,
			req.TableName, req.TimeRange.StartTime, req.TimeRange.EndTime)
		if err != nil {
			return nil, status.New(codes.Internal, err.Error()).Err()
		}
	} else {
		rows, err = conn.Query(ctx,
			`SELECT partition_name, size FROM meta.table_partitions
  		 WHERE table_name = $1`,
			req.TableName)
		if err != nil {
			return nil, status.New(codes.Internal, err.Error()).Err()
		}
	}

	var partitions []*metav1.Partition
	for rows.Next() {
		var name string
		var size int64
		if err := rows.Scan(&name, &size); err != nil {
			return nil, status.New(codes.Internal, err.Error()).Err()
		}

		partitions = append(partitions, &metav1.Partition{
			Name: name,
			Size: size,
		})
	}

	if len(partitions) == 0 {
		return nil, status.New(codes.NotFound, "table not found").Err()
	}

	return &metav1.GetTablePartitionsResponse{
		TableName:  req.TableName,
		Partitions: partitions,
	}, nil
}

func (s *server) RecordNewPartition(ctx context.Context, req *metav1.RecordNewPartitionRequest) (*metav1.RecordNewPartitionResponse, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	err = crdbpgx.ExecuteTx(ctx, conn, pgx.TxOptions{}, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}
		for _, column := range req.GetColumns() {
			batch.Queue(`
			UPSERT INTO meta.table_schemas (tenant_id, table_name, column_name, type, updated_at)
			VALUES ($1, $2, $3, $4, now())
			`, req.GetTenantId(), req.GetTableName(), column.GetName(), column.GetType())
		}

		br := tx.SendBatch(ctx, batch)
		return br.Close()
	})
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	partition := req.GetPartition()
	startTime := time.Unix(partition.GetTimeRange().GetStartTime()/1000, 0).Format(time.RFC3339)
	endTime := time.Unix(partition.GetTimeRange().GetEndTime()/1000, 0).Format(time.RFC3339)
	name := partition.GetName()
	size := partition.GetSize()

	_, err = conn.Exec(ctx, `
	INSERT INTO meta.table_partitions (tenant_id, table_name, start_time, end_time, partition_name, size)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, req.GetTenantId(), req.GetTableName(), startTime, endTime, name, size)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	return &metav1.RecordNewPartitionResponse{}, nil
}

func loadConfig() (*config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func newConnectionPool(cfg *config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=verify-full", cfg.CockroachDBUser, cfg.CockroachDBPassword, cfg.CockroachDBURL, cfg.CockroachDBDatabase)
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalln("error: ", err)
	}

	return pool, nil
}

func newServer(pool *pgxpool.Pool) *server {
	return &server{
		pool: pool,
	}
}

func newGrpcServer(s *server, lifecycle fx.Lifecycle) *grpc.Server {
	grpcServer := grpc.NewServer()
	metav1.RegisterMetaServiceServer(grpcServer, s)
	if reflect {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalln("error: ", err)
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go grpcServer.Serve(lis)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			grpcServer.GracefulStop()
			return nil
		}})

	return grpcServer
}

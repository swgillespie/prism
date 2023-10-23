package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"code.prism.io/go/proto"
)

type config struct {
	CockroachDBUser     string `envconfig:"COCKROACHDB_USER" default:"prism"`
	CockroachDBPassword string `envconfig:"COCKROACHDB_PASSWORD"`
	CockroachDBURL      string `envconfig:"COCKROACHDB_URL"`
	CockroachDBDatabase string `envconfig:"COCKROACHDB_DATABASE" default:"prism"`
}

type server struct {
	proto.UnimplementedMetaServiceServer

	pool *pgxpool.Pool
}

func (s *server) GetTableSchema(ctx context.Context, req *proto.GetTableSchemaRequest) (*proto.GetTableSchemaResponse, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	rows, err := conn.Query(ctx, `
		SELECT column_name, type FROM meta.table_schemas
		WHERE table_name = $1
	`, req.TableName)
	if err != nil {
		return nil, status.New(codes.Internal, err.Error()).Err()
	}

	defer rows.Close()
	var columns []*proto.TableColumn
	for rows.Next() {
		var name string
		var ty int32
		if err := rows.Scan(&name, &ty); err != nil {
			return nil, status.New(codes.Internal, err.Error()).Err()
		}

		columns = append(columns, &proto.TableColumn{
			Name: name,
			Type: proto.ColumnType(ty),
		})
	}

	if len(columns) == 0 {
		return nil, status.New(codes.NotFound, "table not found").Err()
	}

	return &proto.GetTableSchemaResponse{
		TableName: req.TableName,
		Columns:   columns,
	}, nil
}

func (s *server) GetTablePartitions(ctx context.Context, req *proto.GetTablePartitionsRequest) (*proto.GetTablePartitionsResponse, error) {
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

	var partitions []*proto.Partition
	for rows.Next() {
		var name string
		var size int64
		if err := rows.Scan(&name, &size); err != nil {
			return nil, status.New(codes.Internal, err.Error()).Err()
		}

		partitions = append(partitions, &proto.Partition{
			Name: name,
			Size: size,
		})
	}

	if len(partitions) == 0 {
		return nil, status.New(codes.NotFound, "table not found").Err()
	}

	return &proto.GetTablePartitionsResponse{
		TableName:  req.TableName,
		Partitions: partitions,
	}, nil
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
	proto.RegisterMetaServiceServer(grpcServer, s)
	if *reflect {
		reflection.Register(grpcServer)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
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

var (
	port    = flag.Int("port", 8080, "Port to listen on")
	reflect = flag.Bool("reflect", false, "Enable gRPC reflection")
)

func main() {
	flag.Parse()
	fx.New(
		fx.Provide(zap.NewProduction),
		fx.Provide(loadConfig),
		fx.Provide(newConnectionPool),
		fx.Provide(newServer),
		fx.Provide(newGrpcServer),
		fx.Invoke(func(*grpc.Server) {}),
	).Run()
}

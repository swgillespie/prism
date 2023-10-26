package meta

import (
	"context"

	"code.prism.io/go/proto"
)

//go:generate mockgen -source=meta.go -destination=meta_mock.go -package=meta Client

type (
	Client interface {
		GetTableSchema(ctx context.Context, tableName string) (*TableSchema, error)
	}

	TableSchema struct {
		TableName string
		Columns   []*TableColumn
	}

	TableColumn struct {
		Name string
		Type proto.ColumnType
	}

	client struct {
		client proto.MetaServiceClient
	}
)

func NewClient(protoClient proto.MetaServiceClient) Client {
	return &client{
		client: protoClient,
	}
}

func (c *client) GetTableSchema(ctx context.Context, tableName string) (*TableSchema, error) {
	resp, err := c.client.GetTableSchema(ctx, &proto.GetTableSchemaRequest{
		TableName: tableName,
	})
	if err != nil {
		return nil, err
	}

	var columns []*TableColumn
	for _, col := range resp.Columns {
		columns = append(columns, &TableColumn{
			Name: col.Name,
			Type: col.Type,
		})
	}

	return &TableSchema{
		TableName: resp.TableName,
		Columns:   columns,
	}, nil
}

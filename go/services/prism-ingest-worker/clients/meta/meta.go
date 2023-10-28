package meta

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"code.prism.io/go/proto"
	"code.prism.io/go/services/prism-ingest-worker/config"
)

func NewClient(conf *config.Meta) (proto.MetaServiceClient, error) {
	conn, err := grpc.Dial(conf.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return proto.NewMetaServiceClient(conn), nil
}

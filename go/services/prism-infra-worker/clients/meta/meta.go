package meta

import (
	"google.golang.org/grpc"

	"code.prism.io/go/proto"
	"code.prism.io/go/services/prism-infra-worker/config"
)

func NewClient(conf *config.Meta) (proto.MetaServiceClient, error) {
	conn, err := grpc.Dial(conf.Endpoint)
	if err != nil {
		return nil, err
	}

	return proto.NewMetaServiceClient(conn), nil
}

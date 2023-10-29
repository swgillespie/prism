package meta

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"code.prism.io/go/services/prism-ingest-worker/config"
	metav1 "code.prism.io/proto/rpc/gen/go/prism/meta/v1"
)

func NewClient(conf *config.Meta) (metav1.MetaServiceClient, error) {
	conn, err := grpc.Dial(conf.Endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return metav1.NewMetaServiceClient(conn), nil
}

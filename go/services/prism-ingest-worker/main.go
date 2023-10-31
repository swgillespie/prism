package main

import (
	"errors"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"code.prism.io/go/services/prism-ingest-worker/clients/meta"
	"code.prism.io/go/services/prism-ingest-worker/config"
	"code.prism.io/go/services/prism-ingest-worker/workflows/ingest"
	metav1 "code.prism.io/proto/rpc/gen/go/prism/meta/v1"
	ingestv1 "code.prism.io/proto/workflow/gen/prism/ingest/v1"
)

func newConfigProvider() (config.Provider, error) {
	path := os.Getenv("PRISM_INGEST_WORKER_CONFIG")
	if path == "" {
		return nil, errors.New("PRISM_INGEST_WORKER_CONFIG environment variable is required")
	}

	return config.NewYAMLProvider(path)
}

func newMetaClientProvider(metaConfig *config.Meta) ingest.MetaClientProvider {
	return func() (metav1.MetaServiceClient, error) {
		return meta.NewClient(metaConfig)
	}
}

func newCLI(logger *zap.Logger, workflows *ingest.Workflows, activities *ingest.Activities) (*cli.App, error) {
	return ingestv1.NewIngestCli(
		ingestv1.NewIngestCliOptions().
			WithClient(func(ctx *cli.Context) (client.Client, error) {
				return client.Dial(client.Options{
					Logger: NewLogAdaptor(logger),
				})
			}).
			WithWorker(func(ctx *cli.Context, c client.Client) (worker.Worker, error) {
				w := worker.New(c, ingestv1.IngestTaskQueue, worker.Options{})
				ingestv1.RegisterIngestWorkflows(w, workflows)
				ingestv1.RegisterIngestActivities(w, activities)
				return w, nil
			}),
	)
}

func workerMain() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}

	config, err := newConfigProvider()
	if err != nil {
		return err
	}

	wf := ingest.NewWorkflows()
	act := ingest.NewActivities(config.GetIngest(), newMetaClientProvider(config.GetMeta()))
	app, err := newCLI(logger, wf, act)
	if err != nil {
		return err
	}

	return app.Run(os.Args)
}

func main() {
	if err := workerMain(); err != nil {
		log.Fatalln("error: failed to start worker: ", err)
	}
}

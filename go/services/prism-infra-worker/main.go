package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"code.prism.io/go/services/prism-infra-worker/config"
)

var (
	configFile string
	taskQueue  string

	rootCmd = &cobra.Command{
		Use:   "prism-infra-worker",
		Short: "Temporal worker for prism infrastructure",
		Run: func(cmd *cobra.Command, args []string) {
			fx.New(
				fx.Provide(func() (config.Provider, error) { return config.NewYAMLProvider(configFile) }),
				fx.Provide(zap.NewProduction),
				fx.Provide(newWorker),
				fx.Invoke(func(w worker.Worker) {}),
			).Run()
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "configuration file to use")
	rootCmd.Flags().StringVarP(&taskQueue, "task-queue", "t", "", "Temporal task queue to listen on")
}

func newWorker(
	logger *zap.Logger,
	lifecycle fx.Lifecycle,
) (worker.Worker, error) {
	client, err := client.Dial(client.Options{
		Logger: NewLogAdaptor(logger),
	})
	if err != nil {
		return nil, err
	}

	w := worker.New(client, taskQueue, worker.Options{})
	lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			return w.Start()
		},

		OnStop: func(_ context.Context) error {
			w.Stop()
			return nil
		},
	})

	return w, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}
}

package cmd

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	//go:embed migrations
	migrations embed.FS

	migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
	}

	migrateUpCmd = &cobra.Command{
		Use:   "up",
		Short: "Runs all available migrations up",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getMigrate()
			if err != nil {
				fmt.Println("error: ", err)
				os.Exit(1)
			}

			if err := m.Up(); err != nil {
				if err == migrate.ErrNoChange {
					fmt.Println("no migrations to apply")
					os.Exit(0)
				}

				fmt.Println("Failed to apply migrations:")
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println("migrations complete")
		},
	}

	migrateDownCmd = &cobra.Command{
		Use:   "down",
		Short: "Runs all available migrations down",
		Run: func(cmd *cobra.Command, args []string) {
			m, err := getMigrate()
			if err != nil {
				fmt.Println("error: ", err)
				os.Exit(1)
			}

			if err := m.Down(); err != nil {
				if err == migrate.ErrNoChange {
					fmt.Println("no migrations to apply")
					os.Exit(0)
				}

				fmt.Println("Failed to apply migrations:")
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Println("migrations complete")
		},
	}

	_ migrate.Logger = &logAdaptor{}
)

type (
	logAdaptor struct {
		logger *zap.Logger
	}
)

func getMigrate() (*migrate.Migrate, error) {
	var cfg config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return nil, err
	}
	connString := fmt.Sprintf("cockroachdb://%s:%s@%s/%s?sslmode=verify-full", cfg.CockroachDBUser, cfg.CockroachDBPassword, cfg.CockroachDBURL, cfg.CockroachDBDatabase)
	m, err := migrate.NewWithSourceInstance("iofs", source, connString)
	if err != nil {
		return nil, err
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	m.Log = &logAdaptor{logger: logger}
	return m, nil
}

func (l *logAdaptor) Printf(format string, v ...interface{}) {
	l.logger.Info(strings.TrimSpace(fmt.Sprintf(format, v...)))
}

func (l *logAdaptor) Verbose() bool {
	return true
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
}

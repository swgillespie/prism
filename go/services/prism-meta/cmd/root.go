package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "prism-meta",
		Short: "Prism meta service",
		Long:  `Prism meta service, responsible for maintaining information about the schemas and partitions of all tables.`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}
}

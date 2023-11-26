package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "prismctl",
		Short: "Internal CLI tool for interacting with the Prism API",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("error: ", err)
		os.Exit(1)
	}
}

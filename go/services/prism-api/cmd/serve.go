package cmd

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/cobra"

	"code.prism.io/go/services/prism-api/pkg/auth"
	"code.prism.io/go/services/prism-api/pkg/config"
)

var (
	port int

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run the prism-api server",
		Run: func(cmd *cobra.Command, args []string) {
			e := echo.New()
			auth0Config := config.GetAuth0Config()
			e.Use(middleware.Logger())
			e.Use(middleware.Recover())
			e.Use(auth.Auth0TokenMiddleware(auth0Config))
			e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVarP(&port, "port", "p", 1323, "Port to listen on")
}

package entrypoints

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/config"
	"github.com/flyteorg/datacatalog/pkg/rpc/datacatalogservice"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launches the Data Catalog server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()

		// serve a http healthcheck endpoint
		go func() {
			err := datacatalogservice.ServeHTTPHealthCheck(ctx, cfg)
			if err != nil {
				logger.Errorf(ctx, "Unable to serve http", config.GetConfig().GetHTTPHostAddress(), err)
			}
		}()

		return datacatalogservice.ServeInsecure(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

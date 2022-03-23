package entrypoints

import (
	"context"

	"github.com/flyteorg/datacatalog/pkg/config"
	"github.com/flyteorg/datacatalog/pkg/repositories"
	"github.com/spf13/cobra"
)

var serveDummyCmd = &cobra.Command{
	Use:   "serve-dummy",
	Short: "Launches the Data Catalog server without any connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		cfg := config.GetConfig()
		return repositories.ServeDummy(ctx, cfg)
	},
}

func init() {
	RootCmd.AddCommand(serveDummyCmd)
}

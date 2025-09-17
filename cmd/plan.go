package cmd

import (
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/plan"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage plans",
	Long:  `Manage plans`,
}

var planGetCmd = &cobra.Command{
	Use:   "get [plan-id]",
	Short: "Get a specific plan",
	Long:  `Get details for a specific plan by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		detailData := plan.NewDetailData(httpClient, args[0])
		return renderViewpointData(detailData)
	},
}

func init() {
	planCmd.Flags().String("changeset", "", "Changeset name for the plan")
	planCmd.MarkFlagRequired("changeset")

	planCmd.AddCommand(planGetCmd)
}

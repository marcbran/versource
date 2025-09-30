package cmd

import (
	"time"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/rebase"
	"github.com/spf13/cobra"
)

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Manage rebases",
	Long:  `Manage rebases`,
}

var rebaseGetCmd = &cobra.Command{
	Use:   "get [rebase-id]",
	Short: "Get a specific rebase",
	Long:  `Get details for a specific rebase by ID`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		detailData := rebase.NewDetailData(httpClient, changeset, args[0])
		rebaseResp, err := detailData.LoadData()
		if err != nil {
			return err
		}

		waitForCompletion, err := cmd.Flags().GetBool("wait-for-completion")
		if err != nil {
			return err
		}
		if !waitForCompletion || internal.IsTaskCompleted(rebaseResp.State) {
			return renderValue(rebaseResp, func() string {
				return detailData.ResolveData(*rebaseResp)
			})
		}

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				rebaseResp, err = detailData.LoadData()
				if err != nil {
					return err
				}

				if !internal.IsTaskCompleted(rebaseResp.State) {
					continue
				}

				return renderValue(rebaseResp, func() string {
					return detailData.ResolveData(*rebaseResp)
				})
			}
		}
	},
}

var rebaseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rebases",
	Long:  `List all rebases in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return err
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.NewClient(config)
		tableData := rebase.NewTableData(httpClient, changeset)
		return renderTableData(tableData)
	},
}

func init() {
	rebaseGetCmd.Flags().Bool("wait-for-completion", false, "Wait for the rebase to reach a terminal state before returning")
	rebaseGetCmd.Flags().String("changeset", "", "Changeset name (required)")
	rebaseGetCmd.MarkFlagRequired("changeset")
	rebaseListCmd.Flags().String("changeset", "", "Changeset name (optional)")
	rebaseCmd.AddCommand(rebaseGetCmd)
	rebaseCmd.AddCommand(rebaseListCmd)
}

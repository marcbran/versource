package cmd

import (
	"fmt"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Create a new plan for a module",
	Long:  `Create a new plan for a specific module by its ID with branch name`,
	RunE: func(cmd *cobra.Command, args []string) error {
		moduleID, err := cmd.Flags().GetUint("module-id")
		if err != nil {
			return fmt.Errorf("failed to get module-id flag: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		if moduleID == 0 {
			return fmt.Errorf("module-id is required")
		}
		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}

		config, err := LoadConfig()
		if err != nil {
			return err
		}

		client := http.NewClient(config)

		req := internal.CreatePlanRequest{
			ModuleID:  moduleID,
			Changeset: changeset,
		}

		plan, err := client.CreatePlan(cmd.Context(), req)
		if err != nil {
			return err
		}

		fmt.Printf("Plan created successfully with ID: %d\n", plan.ID)
		return nil
	},
}

func init() {
	planCmd.Flags().Uint("module-id", 0, "Module ID")
	planCmd.Flags().String("changeset", "", "Changeset name for the plan")
	planCmd.MarkFlagRequired("module-id")
	planCmd.MarkFlagRequired("changeset")
}

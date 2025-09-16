package cmd

import (
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/spf13/cobra"
)

var componentPlanCmd = &cobra.Command{
	Use:   "plan [component-id]",
	Short: "Create a new plan for a component",
	Long:  `Create a new plan for a specific component by its ID with branch name`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		componentIDStr := args[0]
		componentID, err := strconv.ParseUint(componentIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid component ID: %w", err)
		}

		changeset, err := cmd.Flags().GetString("changeset")
		if err != nil {
			return fmt.Errorf("failed to get changeset flag: %w", err)
		}

		if changeset == "" {
			return fmt.Errorf("changeset is required")
		}

		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}

		client := client.NewClient(config)

		req := internal.CreatePlanRequest{
			ComponentID: uint(componentID),
			Changeset:   changeset,
		}

		plan, err := client.CreatePlan(cmd.Context(), req)
		if err != nil {
			return err
		}

		return formatOutput(plan, "Plan created successfully with ID: %d\n", plan.ID)
	},
}

func init() {
	componentPlanCmd.Flags().String("changeset", "", "Changeset name for the plan")
	componentPlanCmd.MarkFlagRequired("changeset")
}

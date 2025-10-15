package cmd

import (
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/resource"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage resources",
	Long:  `Manage resources`,
}

var resourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resources",
	Long:  `List all resources in the system`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := LoadConfig(cmd)
		if err != nil {
			return err
		}
		httpClient := client.New(config)
		tableData := resource.NewTableData(httpClient)
		return renderTableData(tableData)
	},
}

func init() {
	resourceCmd.AddCommand(resourceListCmd)
}

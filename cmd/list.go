package cmd

import (
	"encoding/json"
	"github.com/marcbran/versource/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all resources of a certain view",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir := os.Getenv("VERSOURCE_DATA_HOME")
		if dataDir == "" {
			dataDir = path.Join(os.Getenv("XDG_DATA_HOME"), "versource")
		}
		view, err := cmd.Flags().GetString("view")
		if err != nil {
			return err
		}
		list, err := internal.List(cmd.Context(), dataDir, view)
		if err != nil {
			return err
		}
		b, err := json.Marshal(list)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	listCmd.Flags().StringP("view", "v", "", "decides which items to list")
}

package cmd

import (
	"github.com/marcbran/versource/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Syncs the resources configured in the current directory",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := os.Getenv("VERSOURCE_CONFIG_HOME")
		if configDir == "" {
			configDir = path.Join(os.Getenv("XDG_CONFIG_HOME"), "versource")
		}
		dataDir := os.Getenv("VERSOURCE_DATA_HOME")
		if dataDir == "" {
			dataDir = path.Join(os.Getenv("XDG_DATA_HOME"), "versource")
		}
		include, err := cmd.Flags().GetStringArray("include")
		if err != nil {
			return err
		}
		exclude, err := cmd.Flags().GetStringArray("exclude")
		if err != nil {
			return err
		}
		return internal.Sync(cmd.Context(), include, exclude, configDir, dataDir)
	},
}

func init() {
	syncCmd.Flags().StringArrayP("include", "i", []string{}, "explicitly includes a group of resources into the sync")
	syncCmd.Flags().StringArrayP("exclude", "e", []string{}, "explicitly excludes a group of resources from the sync")
}

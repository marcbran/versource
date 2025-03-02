package cmd

import (
	"github.com/marcbran/versource/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new resource repository at the current directory",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := os.Getenv("VERSOURCE_CONFIG_HOME")
		if configDir == "" {
			configDir = path.Join(os.Getenv("XDG_CONFIG_HOME"), "versource")
		}
		return internal.Init(cmd.Context(), configDir)
	},
}

package cmd

import (
	"github.com/marcbran/versource/internal"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs a particular action related to the provided resource",
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
		logFile, err := logFile(dataDir)
		if err != nil {
			return err
		}
		defer logFile.Close()

		resource := os.Getenv("resource")
		err = internal.Run(cmd.Context(), configDir, dataDir, resource)
		if err != nil {
			return err
		}
		return nil
	},
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "versource",
	Short: "Versource is a versioned resource manager",
	Long:  ``,
}

func init() {
	rootCmd.AddCommand(changesetCmd)
	rootCmd.AddCommand(componentCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

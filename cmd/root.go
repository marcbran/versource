package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "versource",
	Short: "Versource is a versioned resource manager",
	Long:  ``,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text or json)")
	rootCmd.AddCommand(changesetCmd)
	rootCmd.AddCommand(componentCmd)
	rootCmd.AddCommand(moduleCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
}

func formatOutput(data any, textFormat string, textArgs ...any) error {
	switch outputFormat {
	case "json":
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case "text":
		fmt.Printf(textFormat, textArgs...)
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return nil
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal/tui/platform"
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
	rootCmd.PersistentFlags().String("config", "default", "Configuration key to use (defaults to 'default')")
	rootCmd.AddCommand(changesetCmd)
	rootCmd.AddCommand(componentCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(rebaseCmd)
	rootCmd.AddCommand(moduleCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(uiCmd)
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

func renderViewpointData[T any](detailData platform.ViewportData[T]) error {
	resp, err := detailData.LoadData()
	if err != nil {
		return err
	}
	return renderValue(resp, func() string {
		return detailData.ResolveData(*resp)
	})
}

func renderTableData[T any](tableData platform.TableData[T]) error {
	resp, err := tableData.LoadData()
	if err != nil {
		return err
	}
	return renderValue(resp, func() string {
		columns, rows, _ := tableData.ResolveData(resp)
		return renderTable(columns, rows)
	})
}

func renderValue(data any, textFunc func() string) error {
	switch outputFormat {
	case "json":
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case "text":
		fmt.Print(textFunc())
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}
	return nil
}

func renderTable(columns []table.Column, rows []table.Row) string {
	if len(rows) == 0 {
		return "No data found\n"
	}

	columnWidths := make([]int, len(columns))

	for i, col := range columns {
		columnWidths[i] = len(col.Title)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(columnWidths) && len(cell) > columnWidths[i] {
				columnWidths[i] = len(cell)
			}
		}
	}

	for i := range columnWidths {
		columnWidths[i] += 3
	}

	var result strings.Builder

	var formatParts []string
	for _, width := range columnWidths {
		formatParts = append(formatParts, fmt.Sprintf("%%-%ds", width))
	}

	headerFormat := strings.Join(formatParts, " ")
	headerValues := make([]interface{}, len(columns))
	for i, col := range columns {
		headerValues[i] = strings.ToUpper(col.Title)
	}
	header := fmt.Sprintf(headerFormat, headerValues...)
	result.WriteString(header + "\n")

	formatString := strings.Join(formatParts, " ")
	for _, row := range rows {
		rowValues := make([]interface{}, len(row))
		for i, cell := range row {
			rowValues[i] = cell
		}
		result.WriteString(fmt.Sprintf(formatString+"\n", rowValues...))
	}

	return result.String()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

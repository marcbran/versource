package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(changesetCmd)
	rootCmd.AddCommand(componentCmd)
	rootCmd.AddCommand(mergeCmd)
	rootCmd.AddCommand(rebaseCmd)
	rootCmd.AddCommand(moduleCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(resourceCmd)
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

func renderViewportViewData[T any, V any](detailData platform.ViewportViewData[T, V]) error {
	resp, err := detailData.LoadData()
	if err != nil {
		return err
	}
	return renderViewModel[T, V](*resp, func() V {
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

func renderViewModel[T any, V any](data T, viewModelFunc func() V) error {
	return renderValue(data, func() string {
		viewModel := viewModelFunc()
		yamlData, err := yaml.Marshal(viewModel)
		if err != nil {
			return fmt.Sprintf("Error marshaling to YAML: %v", err)
		}
		return string(yamlData)
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

func waitForTaskCompletion[T any, V any](
	ctx context.Context,
	waitForCompletion bool,
	detailData platform.ViewportViewData[T, V],
	isCompleted func(T) bool,
) error {
	data, err := detailData.LoadData()
	if err != nil {
		return err
	}

	if !waitForCompletion || isCompleted(*data) {
		return renderViewModel(*data, func() V {
			return detailData.ResolveData(*data)
		})
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			data, err = detailData.LoadData()
			if err != nil {
				return err
			}

			if !isCompleted(*data) {
				continue
			}

			return renderViewModel(*data, func() V {
				return detailData.ResolveData(*data)
			})
		}
	}
}

func waitForTableCompletion[T any](
	ctx context.Context,
	waitForCompletion bool,
	tableData platform.TableData[T],
	isCompleted func([]T) bool,
) error {
	data, err := tableData.LoadData()
	if err != nil {
		return err
	}

	if !waitForCompletion || isCompleted(data) {
		return renderValue(data, func() string {
			columns, rows, _ := tableData.ResolveData(data)
			return renderTable(columns, rows)
		})
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			data, err = tableData.LoadData()
			if err != nil {
				return err
			}

			if !isCompleted(data) {
				continue
			}

			return renderValue(data, func() string {
				columns, rows, _ := tableData.ResolveData(data)
				return renderTable(columns, rows)
			})
		}
	}
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

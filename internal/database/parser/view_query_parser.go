package parser

import (
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type SQLViewQueryParser struct{}

func NewSQLViewQueryParser() *SQLViewQueryParser {
	return &SQLViewQueryParser{}
}

func (p *SQLViewQueryParser) Parse(query string) (string, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return "", fmt.Errorf("invalid SQL query: %w", err)
	}

	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return "", fmt.Errorf("query must be a SELECT statement")
	}

	expectedColumns := []string{
		"uuid",
		"provider",
		"provider_alias",
		"resource_type",
		"namespace",
		"name",
		"attributes",
	}

	selectedColumns := make(map[string]bool)
	for _, expr := range selectStmt.SelectExprs {
		switch col := expr.(type) {
		case *sqlparser.StarExpr:
			return "", fmt.Errorf("SELECT * is not allowed, must specify exact columns")
		case *sqlparser.AliasedExpr:
			colName := sqlparser.String(col.Expr)
			selectedColumns[strings.ToLower(colName)] = true
		}
	}

	for _, expectedCol := range expectedColumns {
		if !selectedColumns[expectedCol] {
			return "", fmt.Errorf("missing required column: %s", expectedCol)
		}
	}

	if len(selectedColumns) != len(expectedColumns) {
		return "", fmt.Errorf("query must return exactly %d columns matching the resources table schema", len(expectedColumns))
	}

	return query, nil
}

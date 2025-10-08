package parser

import (
	"fmt"
	"strings"

	"github.com/marcbran/versource/internal"
	"github.com/xwb1989/sqlparser"
)

type SQLViewQueryParser struct{}

func NewSQLViewQueryParser() *SQLViewQueryParser {
	return &SQLViewQueryParser{}
}

func (p *SQLViewQueryParser) Parse(query string) (*internal.ViewResource, error) {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("invalid SQL query: %w", err)
	}

	selectStmt, ok := stmt.(*sqlparser.Select)
	if !ok {
		return nil, fmt.Errorf("query must be a SELECT statement")
	}

	if len(selectStmt.From) == 0 {
		return nil, fmt.Errorf("query must have a FROM clause")
	}

	fromTable, ok := selectStmt.From[0].(*sqlparser.AliasedTableExpr)
	if !ok {
		return nil, fmt.Errorf("FROM clause must reference a table")
	}

	tableName := sqlparser.String(fromTable.Expr)
	if !strings.EqualFold(tableName, "resources") {
		return nil, fmt.Errorf("query must use FROM resources")
	}

	if selectStmt.Where == nil {
		return nil, fmt.Errorf("query must have a WHERE clause")
	}

	whereExpr := selectStmt.Where.Expr
	parentProviderValue, parentResourceTypeValue, err := extractProviderAndResourceTypeFromWhere(whereExpr)
	if err != nil {
		return nil, err
	}

	if parentProviderValue == "" {
		return nil, fmt.Errorf("WHERE clause must include provider = 'string'")
	}

	if parentResourceTypeValue == "" {
		return nil, fmt.Errorf("WHERE clause must include resource_type = 'string'")
	}

	expectedColumns := []string{
		"provider",
		"resource_type",
		"name",
	}

	output := make(map[string]sqlparser.Expr)
	for _, expr := range selectStmt.SelectExprs {
		switch col := expr.(type) {
		case *sqlparser.StarExpr:
			return nil, fmt.Errorf("SELECT * is not allowed, must specify exact columns")
		case *sqlparser.AliasedExpr:
			var outName string
			if !col.As.EqualString("") {
				outName = strings.ToLower(col.As.String())
			} else {
				switch e := col.Expr.(type) {
				case *sqlparser.ColName:
					outName = strings.ToLower(e.Name.String())
				default:
					outName = strings.ToLower(sqlparser.String(col.Expr))
				}
			}
			output[outName] = col.Expr
		}
	}

	for _, expectedCol := range expectedColumns {
		if _, ok := output[expectedCol]; !ok {
			return nil, fmt.Errorf("missing required column: %s", expectedCol)
		}
	}

	if len(output) != len(expectedColumns) {
		return nil, fmt.Errorf("query must return exactly %d columns: provider, resource_type, name", len(expectedColumns))
	}

	var providerValue, resourceTypeValue, nameValue string

	if expr, ok := output["provider"]; ok {
		switch v := expr.(type) {
		case *sqlparser.SQLVal:
			if v.Type != sqlparser.StrVal {
				return nil, fmt.Errorf("provider must be a static string literal")
			}
			providerValue = string(v.Val)
		default:
			return nil, fmt.Errorf("provider must be a static string literal")
		}
	}

	if expr, ok := output["resource_type"]; ok {
		switch v := expr.(type) {
		case *sqlparser.SQLVal:
			if v.Type != sqlparser.StrVal {
				return nil, fmt.Errorf("resource_type must be a static string literal")
			}
			resourceTypeValue = string(v.Val)
		default:
			return nil, fmt.Errorf("resource_type must be a static string literal")
		}
	}

	if expr, ok := output["name"]; ok {
		switch v := expr.(type) {
		case *sqlparser.SQLVal:
			if v.Type != sqlparser.StrVal {
				return nil, fmt.Errorf("name must be a static string literal")
			}
			nameValue = string(v.Val)
		default:
			return nil, fmt.Errorf("name must be a static string literal")
		}
	}

	name := fmt.Sprintf("%s_%s_%s_%s_%s", providerValue, resourceTypeValue, parentProviderValue, parentResourceTypeValue, nameValue)

	return &internal.ViewResource{
		Name:  name,
		Query: query,
	}, nil
}

func extractProviderAndResourceTypeFromWhere(expr sqlparser.Expr) (string, string, error) {
	var providerValue, resourceTypeValue string

	switch e := expr.(type) {
	case *sqlparser.AndExpr:
		leftProvider, leftResourceType, err := extractProviderAndResourceTypeFromWhere(e.Left)
		if err != nil {
			return "", "", err
		}
		rightProvider, rightResourceType, err := extractProviderAndResourceTypeFromWhere(e.Right)
		if err != nil {
			return "", "", err
		}

		if leftProvider != "" {
			providerValue = leftProvider
		}
		if rightProvider != "" {
			providerValue = rightProvider
		}
		if leftResourceType != "" {
			resourceTypeValue = leftResourceType
		}
		if rightResourceType != "" {
			resourceTypeValue = rightResourceType
		}

	case *sqlparser.ComparisonExpr:
		if e.Operator == sqlparser.EqualStr {
			colName, ok := e.Left.(*sqlparser.ColName)
			if !ok {
				return "", "", nil
			}

			columnName := strings.ToLower(colName.Name.String())
			if columnName == "provider" || columnName == "resource_type" {
				val, ok := e.Right.(*sqlparser.SQLVal)
				if !ok || val.Type != sqlparser.StrVal {
					return "", "", fmt.Errorf("%s must be a static string literal in WHERE clause", columnName)
				}

				value := string(val.Val)
				if columnName == "provider" {
					providerValue = value
				} else if columnName == "resource_type" {
					resourceTypeValue = value
				}
			}
		}
	}

	return providerValue, resourceTypeValue, nil
}

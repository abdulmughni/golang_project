package images

import (
	"fmt"
	"strings"
)

type InsertBatch struct {
	TableName string
	Columns   []string
	Values    [][]any
}

func NewInsertBatch(tableName string, columns []string) *InsertBatch {
	return &InsertBatch{
		TableName: tableName,
		Columns:   columns,
		Values:    make([][]any, 0),
	}
}

func (b *InsertBatch) Append(rowValues []any) error {
	if len(rowValues) != len(b.Columns) {
		return fmt.Errorf("number of values does not match number of columns")
	}

	b.Values = append(b.Values, rowValues)
	return nil
}

func (b *InsertBatch) IsEmpty() bool {
	return len(b.Values) == 0
}

func (b *InsertBatch) FinalizeQuery() (string, []any) {
	var placeholders []string
	argIndex := 1

	for range b.Values {
		var rowPlaceholders []string
		for range b.Columns {
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", argIndex))
			argIndex++
		}
		placeholders = append(placeholders, fmt.Sprintf("(%s)", strings.Join(rowPlaceholders, ", ")))
	}

	query := fmt.Sprintf("INSERT INTO st_schema.%s (%s) VALUES %s",
		b.TableName,
		strings.Join(b.Columns, ", "),
		strings.Join(placeholders, ", "))

	var args []any
	for _, row := range b.Values {
		args = append(args, row...)
	}

	return query, args
}

package factory

import (
	"strings"

	"github.com/nauyey/factory/utils"
)

const factoryTag = "factory"

// table represent the map table of factory instance in database.
// It contains name and columns info of the database table.
type table struct {
	name    string
	columns []*column
}

// getPrimaryKeys returns all primary key fields in the table.
func (tbl *table) getPrimaryKeys() []string {
	var keys []string

	for _, col := range tbl.columns {
		if col.isPrimaryKey {
			keys = append(keys, col.name)
		}
	}

	return keys
}

// getPrimaryColumns returns all primary key fields in the table.
func (tbl *table) getPrimaryColumns() []*column {
	var columns []*column

	for _, col := range tbl.columns {
		if col.isPrimaryKey {
			columns = append(columns, col)
		}
	}

	return columns
}

// column defines a table column
type column struct {
	originalModelIndex int
	name               string
	isPrimaryKey       bool
}

// newTable creates a table instance from a Factory instance.
// It does the following:
// Map model struct fields to database table fields by tags declared in the model struct.
// 1. If a struct field, like ID, has tag `factory:"id"`, then the field will be map to be the field "id" in database table.
// 2. If a struct field, like ID, has tag `factory:"id,primary"`, then the field will be map to table field "id",
// and factory will treat it as the primary key of the table.
// 3. If a struct field, like NickName, has tag `factory:""`, `factory:","`, `factory:",primary"` or `factory:",anything else"`, then the field will
// be map to the table field named "nick_name". In this situation, factory just use the snake case of the original struct field name as table field name.
//
// TODO 1: consider query primary key from DB
// TODO 2: consider query auto increment key from DB. Like, select * from COLUMNS where  TABLE_SCHEMA='yourschema' and TABLE_NAME='yourtable' and EXTRA like '%auto_increment%'
func newTable(f *Factory) *table {
	// init table info
	modelType := f.ModelType
	table := &table{
		name: f.Table,
	}

	numField := modelType.NumField()
	for i := 0; i < numField; i++ {
		field := modelType.Field(i)

		tag, ok := field.Tag.Lookup(factoryTag)
		if !ok {
			continue
		}

		columnDesc := utils.StringSliceTrim(strings.Split(tag, ","), " ")
		name := columnDesc[0]
		columnDescExtra := utils.StringSliceToLower(columnDesc[1:])

		if name == "" {
			name = utils.SnakeCase(field.Name)
		}

		table.columns = append(table.columns, &column{
			originalModelIndex: i,
			name:               name,
			isPrimaryKey:       utils.StringSliceContains(columnDescExtra, "primary"),
		})
	}

	return table
}

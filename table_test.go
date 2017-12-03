package factory

import (
	"reflect"
	"testing"
	"time"
)

func TestNewTable(t *testing.T) {
	type testUser struct {
		ID           int64     `factory:"id,primary"`
		Name         string    `factory:"name"`
		NickName     string    `factory:"nick_name"`
		Age          int32     `factory:"age,"`
		FromCountry  string    `factory:""`
		BirthTime    time.Time `factory:","`
		CurrentTime  time.Time `factory:",xxx"`
		NotSaveField string
	}

	// define factory
	userFactory := &Factory{
		ModelType: reflect.TypeOf(testUser{}),
		Table:     "user_table",
		FiledValues: map[string]interface{}{
			"Name": "test name",
		},
	}

	userTable := newTable(userFactory)

	if userTable.name != "user_table" {
		t.Errorf("newTable failed with name=%s, want name=user_table", userTable.name)
	}
	if len(userTable.columns) != 7 {
		t.Errorf("newTable failed with len(columns)=%d, want len(columns)=6", len(userTable.columns))
	}

	expectColumns := []*column{
		&column{name: "id", isPrimaryKey: true, originalModelIndex: 0},
		&column{name: "name", isPrimaryKey: false, originalModelIndex: 1},
		&column{name: "nick_name", isPrimaryKey: false, originalModelIndex: 2},
		&column{name: "age", isPrimaryKey: false, originalModelIndex: 3},
		&column{name: "from_country", isPrimaryKey: false, originalModelIndex: 4},
		&column{name: "birth_time", isPrimaryKey: false, originalModelIndex: 5},
		&column{name: "current_time", isPrimaryKey: false, originalModelIndex: 6},
	}

	for i := 0; i < len(expectColumns); i++ {
		if userTable.columns[i].name != expectColumns[i].name {
			t.Errorf("newTable failed with name=%s, want name=%s",
				userTable.columns[i].name, expectColumns[i].name)
		}
		if userTable.columns[i].isPrimaryKey != expectColumns[i].isPrimaryKey {
			t.Errorf("newTable failed with isPrimaryKey=%v, want isPrimaryKey=%v",
				userTable.columns[i].name, expectColumns[i].isPrimaryKey)
		}
		if userTable.columns[i].originalModelIndex != expectColumns[i].originalModelIndex {
			t.Errorf("newTable failed with originalModelIndex=%d, want originalModelIndex=%d",
				userTable.columns[0].originalModelIndex, expectColumns[i].originalModelIndex)
		}
	}
}

func TestTableMethods(t *testing.T) {
	tbl := table{
		name: "tbl",
		columns: []*column{
			&column{name: "id", isPrimaryKey: true, originalModelIndex: 0},
			&column{name: "name", isPrimaryKey: true, originalModelIndex: 1},
			&column{name: "nick_name", isPrimaryKey: false, originalModelIndex: 2},
			&column{name: "age", isPrimaryKey: false, originalModelIndex: 3},
		},
	}

	primaryKeys := tbl.getPrimaryKeys()
	if len(primaryKeys) != 2 {
		t.Fatalf("getPrimaryKeys failed")
	}
	if primaryKeys[0] != "id" || primaryKeys[1] != "name" {
		t.Errorf("getPrimaryKeys failed")
	}

	primaryColumns := tbl.getPrimaryColumns()
	if len(primaryColumns) != 2 {
		t.Fatalf("getPrimaryColumns failed")
	}
	if primaryColumns[0].name != "id" || primaryColumns[1].name != "name" {
		t.Errorf("getPrimaryColumns failed")
	}
}

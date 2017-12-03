package factory

import "testing"

func TestInsertSQL(t *testing.T) {
	table := "test_table"
	fields := []string{"test_field1", "test_field2", "test_field3"}
	sql := insertSQL(table, fields)
	if sql != `INSERT INTO test_table (test_field1,test_field2,test_field3) VALUES (?,?,?)` {
		t.Errorf("insertSQL failed with sql=%s", sql)
	}
}

func TestSelectSQL(t *testing.T) {
	table := "test_table"
	selectFields := []string{"test_field1", "test_field2", "test_field3"}
	primaryFields := []string{"test_primary_field1"}
	sql := selectSQL(table, selectFields, primaryFields)
	if sql != `SELECT test_field1,test_field2,test_field3 FROM test_table WHERE test_primary_field1=?` {
		t.Errorf("selectSQL failed with sql=%s", sql)
	}
}

func TestDeleteSQL(t *testing.T) {
	table := "test_table"
	primaryFields := []string{"test_primary_field1"}
	sql := deleteSQL(table, primaryFields)
	if sql != `DELETE FROM test_table WHERE test_primary_field1=?` {
		t.Errorf("deleteSQL failed with sql=%s", sql)
	}
}

func TestWhereClause(t *testing.T) {
	fields := []string{}
	sql := whereClause(fields)
	if sql != `` {
		t.Errorf("whereClause failed with sql=%s", sql)
	}

	fields = []string{"test_field1"}
	sql = whereClause(fields)
	if sql != `WHERE test_field1=?` {
		t.Errorf("whereClause failed with sql=%s", sql)
	}

	fields = []string{"test_field1", "test_field2"}
	sql = whereClause(fields)
	if sql != `WHERE test_field1=? AND test_field2=?` {
		t.Errorf("whereClause failed with sql=%s", sql)
	}
}

package dao

import (
	"reflect"
	"testing"
)

// Mock implementation of db.Dbfunc
type mockDbfunc struct{}

func (m *mockDbfunc) TableName() string {
	return "test_table"
}

func TestGetInsertSql(t *testing.T) {
	u := &dbUtils{}
	d := &mockDbfunc{}
	data := map[string]interface{}{
		"column1": "value1",
		"column2": 123,
	}

	sql, args := u.getInsertSql(d, data)
	expectedSql := "INSERT INTO `test_table`(`column1`,`column2`) VALUES(?,?)"
	expectedSql2 := "INSERT INTO `test_table`(`column2`,`column1`) VALUES(?,?)"
	expectedArgs := []interface{}{"value1", 123}
	expectedArgs2 := []interface{}{123, "value1"}

	if !(sql == expectedSql && reflect.DeepEqual(args, expectedArgs)) && !(sql == expectedSql2 && reflect.DeepEqual(args, expectedArgs2)) {
		t.Errorf("sql: %s, args: %v", sql, args)
	}
}

func TestGetUpdateSql(t *testing.T) {
	u := &dbUtils{}
	d := &mockDbfunc{}
	id := uint(1)
	data := map[string]interface{}{
		"column1": "value1",
		"column2": 123,
	}

	sql, args := u.getUpdateSql(d, id, data)
	expectedSql := "UPDATE `test_table` SET `column1` = ?, `column2` = ? WHERE `id` = ?"
	expectedSql2 := "UPDATE `test_table` SET `column2` = ?, `column1` = ? WHERE `id` = ?"
	expectedArgs := []interface{}{"value1", 123, id}
	expectedArgs2 := []interface{}{123, "value1", id}

	if !(sql == expectedSql && reflect.DeepEqual(args, expectedArgs)) && !(sql == expectedSql2 && reflect.DeepEqual(args, expectedArgs2)) {
		t.Errorf("sql: %s, args: %v", sql, args)
	}
}

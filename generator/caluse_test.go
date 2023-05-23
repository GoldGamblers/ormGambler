package generator

import (
	"gamblerORM/log"
	"reflect"
	"testing"
)

func testSelect(t *testing.T) {
	var caluse Clause
	caluse.Set(LIMIT, 3)
	caluse.Set(SELECT, "User", []string{"*"})
	caluse.Set(WHERE, "Name = ?", "Tom")
	caluse.Set(ORDERBY, "Age ASC")
	sql, vars := caluse.Build(SELECT, WHERE, ORDERBY, LIMIT)
	// 检查一下 var 是什么
	log.Info(sql, vars)
	if sql != "SELECT * FROM User WHERE Name = ? ORDER BY Age ASC LIMIT ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{"Tom", 3}) {
		t.Fatal("failed to build SQLVars")
	}
}

func TestClause_Build(t *testing.T) {
	t.Run("select", func(t *testing.T) {
		testSelect(t)
	})
	t.Run("update", func(t *testing.T) {
		testUpdate(t)
	})
	t.Run("delete", func(t *testing.T) {
		testDelete(t)
	})
}

func testUpdate(t *testing.T) {
	var clause Clause
	clause.Set(UPDATE, "User", map[string]interface{}{"Age": 30})
	clause.Set(WHERE, "Name = ?", "Tom")
	sql, vars := clause.Build(UPDATE, WHERE)
	log.Info(sql, vars)
	if sql != "UPDATE User SET Age = ? WHERE Name = ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{30, "Tom"}) {
		t.Fatal("failed to build SQLVars")
	}
}

func testDelete(t *testing.T) {
	var clause Clause
	clause.Set(DELETE, "User")
	clause.Set(WHERE, "Name = ?", "Tom")

	sql, vars := clause.Build(DELETE, WHERE)
	log.Info(sql, vars)
	if sql != "DELETE FROM User WHERE Name = ?" {
		t.Fatal("failed to build SQL")
	}
	if !reflect.DeepEqual(vars, []interface{}{"Tom"}) {
		t.Fatal("failed to build SQLVars")
	}
}

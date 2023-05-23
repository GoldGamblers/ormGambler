package generator

import (
	"fmt"
	"gamblerORM/log"
	"strings"
)

//目标:实现各个子句的生成规则

// 定义了一个 generator 类型,使用接口类型是方便处理不同类型的数据
type generator func(values ...interface{}) (string, []interface{})

// 保存不同关键词对应的处理方法
var generators map[Type]generator

// 初始化
func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

// genBindVars 把一行的数据组合起来，用问号对应原来数据的位置
func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	log.Infof("genBindVars -> vars = %v\n", vars)
	return strings.Join(vars, ", ")
}

// _insert
func _insert(values ...interface{}) (string, []interface{}) {
	// INSERT INTO $tableName ($fields)
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	log.Infof("_insert -> values = %v, tableName = %v, fields = %v\n", values, tableName, fields)
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

// _values
func _values(values ...interface{}) (string, []interface{}) {
	// VALUES ($v1), ($v2), ...
	var bindstr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	// 一行一行的写数据
	for i, value := range values {
		// 一行有几个数据
		v := value.([]interface{})
		if bindstr == "" {
			bindstr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindstr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	log.Infof("_values -> vars = %v, sql.String() = %v\n", vars, sql.String())
	return sql.String(), vars
}

// _select
func _select(values ...interface{}) (string, []interface{}) {
	// SELECT $fields FROM $tableName
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	log.Infof("_select -> values = %v, fields = %v\n", values, fields)
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

// _limit
func _limit(values ...interface{}) (string, []interface{}) {
	// LIMIT $num
	log.Infof("_limit -> values = %v\n", values)
	return "LIMIT ?", values
}

// _where
func _where(values ...interface{}) (string, []interface{}) {
	// WHERE $desc
	desc, vars := values[0], values[1:]
	log.Infof("_where -> values = %v, desc = %v, vars = %v\n", values, desc, vars)
	return fmt.Sprintf("WHERE %s", desc), vars
}

// _orderBy
func _orderBy(values ...interface{}) (string, []interface{}) {
	log.Infof("_orderBy -> values = %v\n", values)
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

// _update 第一个参数是表名(table)，第二个参数是 map 类型，表示待更新的键值对
func _update(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range m {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	log.Infof("_update -> keys = %v, vars = %v\n", keys, vars)
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ", ")), vars
}

// _delete 只有一个入参，即表名
func _delete(values ...interface{}) (string, []interface{}) {
	log.Infof("_delete -> values = %v\n", values)
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

// _count 只有一个入参，即表名，并复用了 _select 生成器
func _count(values ...interface{}) (string, []interface{}) {
	log.Infof("_count -> values = %v\n", values)
	return _select(values[0], []string{"count(*)"})
}

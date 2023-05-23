package session

import (
	"fmt"
	"gamblerORM/log"
	"gamblerORM/schema"
	"reflect"
	"strings"
)

// 用于放置操作数据库表相关的代码

// Model 用于给 refTable 赋值， refTable 保存解析结果
func (s *Session) Model(value interface{}) *Session {
	// 解析操作比较耗时，所以要把解析的结果保存到 refTable 中，如果传入 Moder 的结构体名称不变，那么就不会更新 refTable 的值
	if s.refTable == nil || reflect.TypeOf(value) != reflect.TypeOf(s.refTable.Model) {
		// 保存解析结果，这个结果是一张表的信息，是 schema 结构的
		s.refTable = schema.Parse(value, s.dialect)
	}
	return s
}

// RefTable 返回 refTable 的值
func (s *Session) RefTable() *schema.Schema {
	// 如果没有被赋值则打印错误日志
	if s.refTable == nil {
		log.Error("Model is not set")
	}
	return s.refTable
}

// CreateTable 创建表
func (s *Session) CreateTable() error {
	// table 是解析结果，是 schema 结构体的形式
	table := s.RefTable()
	// 列信息
	var columns []string
	// 拿到 Fields 里面的 Field 并追加到 列信息里面
	for _, field := range table.Fields {
		columns = append(columns, fmt.Sprintf("%s %s %s", field.Name, field.Type, field.Tag))
	}
	// 用 , 来连接 每一对 field.Name field.Type field.Tag 的值
	desc := strings.Join(columns, ",")
	// 创建表的实际操作
	_, err := s.Raw(fmt.Sprintf("CREATE TABLE %s (%s);", table.Name, desc)).Exec()
	return err
}

// DropTable 删除表
func (s *Session) DropTable() error {
	// s.RefTable() 是 解析后的 schema 结构的结果，其中 Name 字段是表名
	_, err := s.Raw(fmt.Sprintf("DROP TABLE IF EXISTS %s", s.RefTable().Name)).Exec()
	return err
}

// JudgeTableExist 判断表是否存在
func (s *Session) JudgeTableExist() bool {
	// 拿到判断表名是s.RefTable().Name的表是否存在的对应数据库的 sql 语句
	sql, values := s.dialect.TableExistSQL(s.RefTable().Name)
	// 拿到 SQL 语句的执行结果，这个结果是表名 或者 返回一个 err
	res := s.Raw(sql, values...).QueryRow()
	var temp string
	// 获取一行的结果后，调用Scan方法来将返回结果赋值给对象或者结构体
	_ = res.Scan(&temp)
	return temp == s.RefTable().Name
}

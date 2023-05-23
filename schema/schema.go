package schema

import (
	"gamblerORM/dialect"
	"gamblerORM/log"
	"go/ast"
	"reflect"
)

// 目标：实现 ORM 框架中最为核心的转换——对象(object)和表(table)的转换
// 也就是给定一个任意的对象，转换为关系型数据库中的表结构

//数据库表的构成要素：
// 1、表名：对应结构体名
// 2、字段名和类型：对应成员变量和类型
// 3、约束条件(主键、非空等)：对应成员变量的 tag， Java 和 python 等是通过注解来实现的

//eg: schema 语句对应如下结构体：CREATE TABLE `User` (`Name` text PRIMARY KEY, `Age` integer);
//type User struct {
//	Name string `liup:"PRIMARY KEY"`
//	Age  int
//}

// Field 代表数据库的一列的信息（不是数据）
type Field struct {
	Name string
	Type string
	Tag  string
}

// Schema 代表数据库的一张表的信息（不是数据）, 需要把其他对象构建成 schema 的样子
type Schema struct {
	Model      interface{}       // 被映射的对象
	Name       string            //表名
	Fields     []*Field          // 多个列
	FieldNames []string          // 每个列的列名
	fieldMap   map[string]*Field //存储列的信息，也就是 Field
}

// GetField 返回列信息 field，用于测试
func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// RecordValues 返回 dest 对象的字段值，根据数据库中列的顺序，从对象中找到对应的值，按顺序平铺
//INSERT 对应的 SQL 语句一般是这样的：
//
//	INSERT INTO table_name(col1, col2, col3, ...) VALUES
//	(A1, A2, A3, ...),
//	(B1, B2, B3, ...),
//	...
// RecordValues
func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}

type ITableName interface {
	TableName() string
}

// Parse 将任意对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// reflect.Indirect(v)函数用于获取v指向的值,如果v是nil指针，则Indirect返回零值。如果v不是指针，则Indirect返回v
	// dest 是一个对象，例如 &User{} 结构体，使用 reflect.ValueOf() 可以拿到 User 结构体里面每个字段的值，再使用 type 拿到每个字段的类型，最后的 .Type() 是获取类型的，如 main.User
	// 整体最后返回的是一个指针类型, 需要 reflect.Indirect() 获取指针指向的实例
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	//fmt.Printf("Debug msg : schema.go -> modelType =  %v\n", modelType)
	log.Infof("Parse -> modelType =  %v\n", modelType)

	var tableName string
	t, ok := dest.(ITableName)
	if !ok {
		tableName = modelType.Name()
	} else {
		tableName = t.TableName()
	}

	schema := &Schema{
		Model:    dest,                    // 结构体
		Name:     tableName,               // 例如 User, 作为表名
		fieldMap: make(map[string]*Field), // 建立映射
	}
	//  modelType 里面是 User 结构体里面每个字段的数据，NumField() 获取字段的数量
	for i := 0; i < modelType.NumField(); i++ {
		// 拿到每一个字段值
		p := modelType.Field(i)
		// 判断条件没懂
		if !p.Anonymous && ast.IsExported(p.Name) {
			// p.Name 即字段名，p.Type 即字段类型了，p.Tag 即额外的约束条件
			field := &Field{
				Name: p.Name,                                              // 字段名
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))), // 类型
			}
			// 设置 field 的 tag 值,参数是 tag 的 key 值
			if v, ok := p.Tag.Lookup("gamblerORM"); ok {
				field.Tag = v
			}
			// 一个 field 是一个列的信息，把每个列添加到 schema 中
			schema.Fields = append(schema.Fields, field)
			// 把每个字段名添加到字段名列表中，保存所有的列名
			schema.FieldNames = append(schema.FieldNames, p.Name)
			// 将列名和列的信息对应起来，列的信息包括 Field 里面的信息
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

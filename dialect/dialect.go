package dialect

import "reflect"

// 主要目的是使用 dialect 隔离不同数据库之间的差异，便于扩展，实现了一些特定的 SQL 语句的转换
// 1、映射数据结构，如 Go 语言中的 int、int8、int16 等类型均对应 SQLite 中的 integer 类型
// 2、不同的数据库之间有差异，需要提取差异，让代码能够在不同的数据库之间实现复用

// 反射：程序在运行时，通过检查其定义的变量以及值，进而找到其对应的真实类型
// 意义：对于query函数来说为了适配任意类型的struct，通过在运行时检查传递的struct的具体类型并根据其包含的字段来生成对应的sql
// 具体类型可以使用reflect.Type 类型来表示, reflect.TypeOf() 方法可以获取类型的具体指
// 具体的值则可使用reflect.Value类型来表示, reflect.ValueOf() 方法可以获取具体值
// 实际的类型可以用 .Kind() 获取，如struct
// NumField()方法获取一个struct所有的fields
// Field(i int)获取指定第 i 个 field 的 reflect.Value 值
// Int()和String()主要用于从 reflect.Value 提取对应值作为 int64 和 string 类型
// reflect.SliceOf() 某个数据类型的切片类型

// 对应于不同的数据库
var dialectMap = map[string]Dialect{}

// Dialect 接口用于扩展到其他数据库，编写的支持其他数据库的文件需要实现这两个方法
type Dialect interface {
	DataTypeOf(typ reflect.Value) string                    // 用于将 Go 语言的类型转换为数据库的数据类型
	TableExistSQL(tableName string) (string, []interface{}) //返回某个表是否存在的 SQL 语句
}

// RegisterDialect 注册 dialect 实例
func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}

// GetDialect 获取 dialect实例
func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectMap[name]
	return
}

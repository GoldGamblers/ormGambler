package dialect

import (
	"fmt"
	"reflect"
	"time"
)

type sqlite3 struct{}

// init 包在第一次加载时，会将 sqlite3 的 dialect 自动注册到全局
func init() {
	RegisterDialect("sqlite3", &sqlite3{})
}

// DataTypeOf 用于将 Go 语言的类型转换为 sqlite3 数据库的数据类型
func (s *sqlite3) DataTypeOf(typ reflect.Value) string {
	//TODO implement me
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	// 64位计算机 int 是 8 字节， 32位计算机 int 是 4字节
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
		return "integer"
	case reflect.Int64, reflect.Uint64:
		return "bigint"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.String:
		return "text"
	case reflect.Array, reflect.Slice:
		return "blob"
	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("sqlite3.go : invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

// TableExistSQL  返回在 SQLite 中判断表 tableName 是否存在的 SQL 语句
func (s *sqlite3) TableExistSQL(tableName string) (string, []interface{}) {
	//TODO implement me
	args := []interface{}{tableName}
	return "SELECT name FROM sqlite_master WHERE type='table' and name = ?", args
}

// 通过如下检测确保某个类型实现了某个接口的所有方法
// 注释：将空值 nil 转换为 *sqlite3 类型，再转换为 Dialect 接口，如果转换失败，说明 sqlite3 并没有实现 Dialect 接口的所有方法
var _ Dialect = (*sqlite3)(nil)

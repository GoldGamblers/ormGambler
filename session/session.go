package session

import (
	"database/sql"
	"gamblerORM/dialect"
	"gamblerORM/generator"
	"gamblerORM/log"
	"gamblerORM/schema"
	"strings"
)

// Session 用于实现与数据库的交互

type Session struct {
	db       *sql.DB          // 使用 sql.Open() 方法连接数据库成功之后返回的指针
	sql      strings.Builder  // 拼接 SQL 语句,调用 Raw() 方法即可改变以下两个变量的值
	sqlVars  []interface{}    // SQL 语句中占位符的对应值
	dialect  dialect.Dialect  // 存储对不同数据库的匹配
	refTable *schema.Schema   // 代表一张表的信息
	clause   generator.Clause // 添加 clause 用于拼接字符串
	tx       *sql.Tx          // 添加对事务的支持，使用 tx 来实现事务
}

// CommonDB 定义一个集合，用于实现 事务方式使用数据库
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// New 创建一个新的 Session 用来操作数据库，Session中有db操作句柄和对不同数据库的适配
func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

// Clear 清空 sql 和 slqVars,使得session可以被复用，开启一次会话可以多次sql
func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	s.clause = generator.Clause{}
}

// 用于检查这两种使用数据库的方式中，是否全部实现了接口要求的方法
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

// DB 返回 Session 的 db, 也就是使用 sql.Open() 方法连接数据库成功之后返回的指针
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

// Raw 用来改变 Session 中的 sql 和 sqlVars 字段，这两个字符用来拼接 sql 语句
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

// Exec 封装 sql 的Exec()方法，可以统一打印日志和清除sql语句
func (s *Session) Exec() (result sql.Result, err error) {
	// 使用完毕后关闭数据库连接
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		// log.go 中定义 Error = errorLog.Println
		log.Error(err)
	}
	return
}

//QueryRow 封装 sql 的 QueryRow 方法，从数据库中获取一条数据
func (s *Session) QueryRow() *sql.Row {
	//执行查询之前先清空 sql
	defer s.Clear()
	// log.go 中定义 Info = infoLog.Println
	log.Info(s.sql.String(), s.sqlVars)
	// 实际执行
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

// QueryRows 封装 sql 的 Query 方法，从数据库中获取多条数据
func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	//执行查询之前先清空 sql
	defer s.Clear()
	// log.go 中定义 Info = infoLog.Println
	log.Info(s.sql.String(), s.sqlVars)
	// 实际执行
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

package gamblerORM

import (
	"errors"
	"gamblerORM/session"
	"reflect"
	"testing"
)
import _ "github.com/mattn/go-sqlite3"

type User struct {
	Name string `gamblerORM:"PRIMARY KEY"`
	Age  int
}

func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("sqlite3", "gamblerORM.db")
	if err != nil {
		t.Fatal("failed to connect", err)
	}
	return engine
}

func TestNewEngine(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("RoolBack", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("Commit", func(t *testing.T) {
		transactionCommit(t)
	})
}

// transactionRollback 执行成功，则会创建一张表 User，并插入一条记录。
func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	// 开启事务， 在事务中创建表并新增一条数据
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Liup", 24})
		// 故意返回了一个自定义 error，最终事务回滚，表创建失败
		return nil, errors.New("ERROR")
	})
	// 如果回滚之后表还存在或者发生了错误，则回滚失败
	if err == nil || s.JudgeTableExist() {
		t.Fatal("Failed to RollBack")
	}
}

// transactionCommit 创建表和插入记录均成功执行，最终通过 s.First() 方法查询到插入的记录
func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	// 创建会话
	s := engine.NewSession()
	// 先删除表
	_ = s.Model(&User{}).DropTable()
	// 开启事务
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Liup", 24})
		return
	})
	u := &User{}
	_ = s.First(u)
	if err != nil || u.Name != "Liup" {
		t.Fatal("Failed to Commit")
	}
}

func TestEngine_Migrate(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text PRIMARY KEY, needDel integer);").Exec()
	_, _ = s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	// 这里的 User 是要合并的新对象
	engine.Migrate(&User{})

	rows, _ := s.Raw("SELECT * FROM User").QueryRows()
	columns, _ := rows.Columns()
	if !reflect.DeepEqual(columns, []string{"Name", "Age"}) {
		t.Fatal("Failed to migrate table User, got columns", columns)
	}
}

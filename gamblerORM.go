package gamblerORM

import (
	"database/sql"
	"fmt"
	"gamblerORM/dialect"
	"gamblerORM/log"
	"gamblerORM/session"
	"strings"
)

type Engine struct {
	db      *sql.DB         // 数据库句柄
	dialect dialect.Dialect // 添加 dialect 实现对不同数据库的支持
}

type TxFunc func(*session.Session) (interface{}, error)

// NewEngine 创建数据库引擎
func NewEngine(driver, source string) (e *Engine, err error) {
	// 连接数据库
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// 发送一个 ping 来确认数据库连接
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// 确认 使用的数据库 对应的 dialect 存在
	dialect, ok := dialect.GetDialect(driver)
	if !ok {
		log.Error("Dialect %s Not Found", driver)
	}
	// 实例化引擎
	e = &Engine{
		db:      db,
		dialect: dialect,
	}
	log.Info("Connection database success")
	return
}

// Close 关闭数据库连接
func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

// NewSession 创建新会话,会话中返回一个数据库的引擎
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

// Transaction 提供对封装的事务方法的调用
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	// 调用 defer 结束事务
	defer func() {
		if p := recover(); p != nil {
			// 发生了 panic，进行回滚
			_ = s.RollBack()
			// 回滚后重新抛出 panic
			panic(p)
		} else if err != nil {
			// 非空的错误，回滚
			_ = s.RollBack()
		} else {
			// 没有发生错误
			err = s.Commit()
		}
	}()
	return f(s)
}

// getDifference 获取两张表的差集，new - old = new field
func getDifference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

// Migrate 实现数据库表的合并
func (engine *Engine) Migrate(value interface{}) error {
	// 使用事务来实现表的合并
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		// 如果要合并的这张表没有同名的原始表
		if !s.Model(value).JudgeTableExist() {
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, err
		}
		// table 是要合并的表，也就是新表，这个是从传入的 value 对象解析得到的
		table := s.RefTable()
		log.Infof("Migrate -> new table columns = %v\n", table.FieldNames)
		// 拿到旧表的所有列，这个是从数据库中查询到的，因为新表不在数据库中，新表目前只是一个对象
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		log.Infof("Migrate -> old table columns = %v\n", columns)
		// 得到差集， 新 - 旧 就是新增的，旧 - 新 就是要删除的
		addCols := getDifference(table.FieldNames, columns)
		delCols := getDifference(columns, table.FieldNames)
		log.Infof("Add cols %v, Delete cols %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}
		if len(addCols) == 0 {
			return
		}
		// 构建新表
		temp := "temp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s FROM %s;", temp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", temp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}

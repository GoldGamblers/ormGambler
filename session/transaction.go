package session

import (
	"gamblerORM/log"
)

// 封装事务的 begin、commit、rollback 方法

// Begin 封装事务的Begin方法
func (s *Session) Begin() (err error) {
	log.Info("Transaction Begin")
	// 调用 s.db.Begin() 得到 *sql.Tx 对象，赋值给 s.tx
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

// Commit 封装事务的Commit方法
func (s *Session) Commit() (err error) {
	log.Info("Transaction Commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
		return
	}
	return
}

// RollBack 封装事务的RollBack方法
func (s *Session) RollBack() (err error) {
	log.Info("Transaction RollBack")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
		return
	}
	return
}

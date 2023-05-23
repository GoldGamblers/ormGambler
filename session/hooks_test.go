package session

import (
	"gamblerORM/log"
	"testing"
)

type Account struct {
	ID       int `gamblerORM:"PRIMARY KEY"`
	Password string
}

// 针对account对象设计的钩子方法
func (account *Account) BeforeInsert(s *Session) error {
	log.Info("Before Insert", account)
	account.ID += 1000
	return nil
}

// 针对account对象设计的钩子方法
func (account *Account) AfterQuery(s *Session) error {
	log.Info("After Query", account)
	account.Password = "******"
	return nil
}

func TestSession_CallMethod(t *testing.T) {
	s := NewSession().Model(&Account{})
	_ = s.DropTable()
	_ = s.CreateTable()
	_, _ = s.Insert(&Account{1, "123456"}, &Account{2, "qwerty"})

	u := &Account{}

	err := s.First(u)
	if err != nil || u.ID != 1001 || u.Password != "******" {
		t.Fatal("Failed to call hooks after query, got", u)
	}
}

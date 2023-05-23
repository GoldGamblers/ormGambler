package session

import (
	"testing"
)

type User struct {
	Name string `gamblerORM:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	// 创建session并解析 User 结构体
	s := NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	if !s.JudgeTableExist() {
		t.Fatal("Failed to create table User")
	}
}

func TestSession_Model(t *testing.T) {
	s := NewSession().Model(&User{})
	table := s.RefTable()
	s.Model(&Session{})
	if table.Name != "User" || s.RefTable().Name != "Session" {
		t.Fatal("Failed to change model")
	}
}

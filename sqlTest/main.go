package main

import (
	"fmt"
	"gamblerORM"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	engine, _ := gamblerORM.NewEngine("sqlite3", "gamblerORM.db")
	//  关闭连接
	defer engine.Close()
	// 创建新会话
	s := engine.NewSession()
	// Raw 返回的是一个 session，所以可以直接继续调用 Exec
	_, _ = s.Raw("DROP TABLE IF EXISTS User;").Exec()
	// 故意多写一遍创建表，测试 log是否正常工作
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	_, _ = s.Raw("CREATE TABLE User(Name text);").Exec()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	// affected 指的是在执行数据库操作(例如插入、更新或删除)后影响到的行数
	fmt.Printf("Exec success, %d affected\n", count)
}

package session

import (
	"gamblerORM/log"
	"reflect"
)

// 定义钩子常量
const (
	BeforeQuery  = "BeforeQuery"
	AfterQuery   = "AfterQuery"
	BeforeUpdate = "BeforeUpdate"
	AfterUpdate  = "AfterUpdate"
	BeforeDelete = "BeforeDelete"
	AfterDelete  = "AfterDelete"
	BeforeInsert = "BeforeInsert"
	AfterInsert  = "AfterInsert"
)

// CallMethod 调用已经注册的钩子
func (s *Session) CallMethod(method string, value interface{}) {
	// s.RefTable().Model 或 value 即当前会话正在操作的对象
	// MethodByName 方法反射得到该对象的钩子方法， method 是 和给对象编写的钩子方法同名的字符串
	fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	//fmt.Printf("CallMethod fm = : %s\n", fm)
	if value != nil {
		fm = reflect.ValueOf(value).MethodByName(method)
	}
	// 将 s *Session 作为入参调用。每一个钩子的入参类型均是 *Session
	param := []reflect.Value{reflect.ValueOf(s)}
	//fmt.Printf("CallMethod param = : %v\n", param)
	if fm.IsValid() {
		if v := fm.Call(param); len(v) > 0 {
			if err, ok := v[0].Interface().(error); ok {
				log.Error(err)
			}
		}
	}
	// 将 CallMethod() 方法在 Find、Insert、Update、Delete 方法内部调用即可
	return
}

package session

import (
	"errors"
	"gamblerORM/generator"
	"reflect"
)

// Insert 实现 insert 功能
// 1）多次调用 clause.Set() 构造好每一个子句。
// 2）调用一次 clause.Build() 按照传入的顺序构造出最终的 SQL 语句。
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	//例如要执行这样的插入语句
	//INSERT INTO table_name(col1, col2, col3, ...) VALUES
	//(A1, A2, A3, ...),
	//(B1, B2, B3, ...),
	//...
	for _, value := range values {
		// 调用钩子 BeforeInsert，但是没有进行测试
		s.CallMethod(BeforeInsert, value)
		table := s.Model(value).RefTable()
		s.clause.Set(generator.INSERT, table.Name, table.FieldNames)
		// 得到和列名对应的一行数据，如有3列，则对应 {A1, B1, C1}
		recordValues = append(recordValues, table.RecordValues(value))
	}
	// 拼接所有的参数得到values子句, recordValues 不只是一条，需要加 ...
	s.clause.Set(generator.VALUES, recordValues...)
	// 按顺序调用 select 子句 和 values 子句
	sql, vars := s.clause.Build(generator.INSERT, generator.VALUES)
	// 执行查询
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	// 调用钩子 AfterInsert
	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}

// Find 实现 Find 功能
// Find 功能的难点和 Insert 恰好反了过来。Insert 需要将已经存在的对象的每一个字段的值平铺开来，而 Find 则是需要根据平铺开的字段的值构造出对象
func (s *Session) Find(values interface{}) error {
	// 调用钩子 BeforeQuery
	s.CallMethod(BeforeQuery, nil)
	// 拿到多个对象的每个字段的值
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	// 获取切片的单个元素的类型 destType
	destType := destSlice.Type().Elem()
	// reflect.New() 方法创建一个 destType 的实例，作为 Model() 的入参，映射出表结构 RefTable()
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	//开始构建子句
	s.clause.Set(generator.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(generator.SELECT, generator.WHERE, generator.ORDERBY, generator.LIMIT)
	// 执行查找
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}
	// 数据库返回的多条记录要使用 Next()来遍历
	for rows.Next() {
		// 遍历每一行记录，利用反射创建 destType 的实例 dest，将 dest 的所有字段平铺开，构造切片 values
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// 调用 rows.Scan() 将该行记录每一列的值依次赋值给 values 中的每一个字段
		if err := rows.Scan(values...); err != nil {
			return err
		}
		// Addr() 方法询问 reflect.Value 变量是否可寻址
		// 调用钩子 AfterQuery
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		// 将 dest 添加到切片 destSlice 中。循环直到所有的记录都添加到切片 destSlice 中
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

// Update 功能实现：kv是多个不定长度的参数
func (s *Session) Update(kv ...interface{}) (int64, error) {
	// 调用钩子 BeforeUpdate
	s.CallMethod(BeforeUpdate, nil)
	// 类相转化
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		// 步长为2，因为是k和v
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}
	// 构造子句, UPDATE 语句，表名和参数
	s.clause.Set(generator.UPDATE, s.RefTable().Name, m)
	// 合成完成的sql语句
	sql, vars := s.clause.Build(generator.UPDATE, generator.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	// 调用钩子 AfterUpdate
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

// Delete 删除功能实现
func (s *Session) Delete() (int64, error) {
	s.clause.Set(generator.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(generator.DELETE, generator.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Count 计数功能实现
func (s *Session) Count() (int64, error) {
	// 调用钩子 BeforeDelete
	s.CallMethod(BeforeDelete, nil)
	// 构造子句
	s.clause.Set(generator.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(generator.COUNT, generator.WHERE)
	// 最终的结果只是一条数据不是多条
	row := s.Raw(sql, vars...).QueryRow()
	var temp int64
	//result, err := s.Raw(sql, vars...).Exec()
	// 把结果复制给 temp
	if err := row.Scan(&temp); err != nil {
		return 0, err
	}
	// 调用钩子 AfterDelete
	s.CallMethod(AfterDelete, nil)
	return temp, nil
}

// Limit 方法实现链式调用，关键是返回 *Session
func (s *Session) Limit(num int) *Session {
	s.clause.Set(generator.LIMIT, num)
	return s
}

// Where 方法实现链式调用，关键是返回 *Session
func (s *Session) Where(desc string, args ...interface{}) *Session {
	var vars []interface{}
	s.clause.Set(generator.WHERE, append(append(vars, desc), args...)...)
	return s
}

// OrderBy 方法实现链式调用，关键是返回 *Session
func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(generator.ORDERBY, desc)
	return s
}

// First 方法实现只返回一条结果, 根据传入的类型，利用反射构造切片，调用 Limit(1) 限制返回的行数，调用 Find 方法获取到查询结果。
func (s *Session) First(value interface{}) error {
	// 通过反射拿到对象值
	dest := reflect.Indirect(reflect.ValueOf(value))
	// reflect.SliceOf 某种数据类型的切片类型，通过反射拿到的类型来创建新的切片
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	// Addr() 方法询问 reflect.Value 变量是否可寻址
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return nil
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}

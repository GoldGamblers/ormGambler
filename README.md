# 对象关系映射 ORM

- [对象关系映射 ORM](#对象关系映射-orm)
  - [一、实现日志分级功能：](#一实现日志分级功能)
  - [二、实现 Session 用于实现和数据库的交互功能](#二实现-session-用于实现和数据库的交互功能)
    - [1、Session封装了三个原生的数据库操作方法：](#1session封装了三个原生的数据库操作方法)
    - [2、实现对对不同数据库的sql语句的支持 (Dialect)](#2实现对对不同数据库的sql语句的支持-dialect)
    - [3、实现对象和表的转换 (schema)](#3实现对象和表的转换-schema)
    - [4、实现新增和查询](#4实现新增和查询)
    - [5、实现修改和删除以及统计功能](#5实现修改和删除以及统计功能)
    - [6、实现钩子功能](#6实现钩子功能)
    - [7、实现事务功能](#7实现事务功能)
  - [三、实现 Engine 用于 Session 交互前的准备工作](#三实现-engine-用于-session-交互前的准备工作)
    - [1、创建 Engine 用于准备工作或者全局的一些工作，包括连接数据库等](#1创建-engine-用于准备工作或者全局的一些工作包括连接数据库等)
    - [2、Engine 提供用于创建对应的 Session 的方法](#2engine-提供用于创建对应的-session-的方法)
    - [3、Engine 提供对封装的事务方法的调用，为用户提供一键式使用的接口](#3engine-提供对封装的事务方法的调用为用户提供一键式使用的接口)
    - [4、Engine 提供对数据库迁移的调用，使用了之前的事务](#4engine-提供对数据库迁移的调用使用了之前的事务)


目的：通过使用描述对象和数据库之间映射的元数据，将面向对象语言程序中的对象自动持久化到关系数据库中。

ORM 框架相当于对象和数据库中间的一个桥梁，借助 ORM 可以避免写繁琐的 SQL 语言，仅仅通过操作具体的对象，就能够完成对关系型数据库的操作。

实现 ORM 的关键技术是反射技术。通过反射，可以获取到对象对应的结构体名称，成员变量、方法等信息。

一般地映射关系：

| 数据库               | 面向对象的编程语言        |
|-------------------|------------------|
| 表(table)          | 类(class/struct)  |
| 记录(record, row)   | 对象(object)       |
| 字段(field, column) | 对象属性(attribute)  |


## 一、实现日志分级功能：

    ·日志分级
    ·不同日志分级用不同颜色进行区分
    ·显示打印日志代码的文件名及行号
使用 log.New() 设置日志信息的格式并重命名，使用 SetOutput() 指定输出的位置。使用 iota 定义日志等级常量，提供日志等级设置函数，实现只显示高于或等于当前日志的日志信息。具体为将小于这个等级的日志在 SetOutput() 方法中使用 ioutil.Discard 参数将日志设置为不显示。


```go
func SetLevel(level int) {
    // 上锁
    mu.Lock()
    // 执行完毕后解锁
    defer mu.Unlock()

    for _, logger := range loggers {
        logger.SetOutput(os.Stdout)
    }

    //如果设置为 ErrorLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志
    if ErrorLevel < level {
        errorLog.SetOutput(ioutil.Discard)
    }

    //如果设置为 InfoLevel，infoLog 的输出会被定向到 ioutil.Discard，即不打印该日志
    if InfoLevel < level {
        infoLog.SetOutput(ioutil.Discard)
    }
}
```


## 二、实现 Session 用于实现和数据库的交互功能
Session 的成员变量如下：

```go
type Session struct {
    db       *sql.DB          // 使用 sql.Open() 方法连接数据库成功之后返回的指针
    sql      strings.Builder  // 拼接 SQL 语句,调用 Raw() 方法即可改变以下两个变量的值
    sqlVars  []interface{}    // SQL 语句中占位符的对应值
    dialect  dialect.Dialect  // 存储对不同数据库的匹配
    refTable *schema.Schema   // 代表一张表的信息
    clause   generator.Clause // 添加 clause 用于拼接字符串
    tx       *sql.Tx          // 添加对事务的支持，使用 tx 来实现事务
}
```

### 1、Session封装了三个原生的数据库操作方法：

```go
Exec()
Query()
QueryRow()
```
目的是为了统一打印日志、执行完成后清空 Session 的 SQL 语句拼接变量以实现对 Session 的复用。
### 2、实现对对不同数据库的sql语句的支持 (Dialect)
数据库中存储的数据的格式和各种编程语言中的类型是有区别的，除此之外不同的数据库支持的数据类型也是有区别的，利用 dialect 也就是方言来提取差异，实现对用户屏蔽不同数据库的差异。

首先定义 Dialect 接口，其他的数据库必须要实现这个接口

```go
type Dialect interface {
    // 将 go 语言数据类型映射为对应的数据库的数据类型
    DataTypeOf(typ reflect.Value) string
    // 检查是否存在表的sql语句
    TableExistSQL(tableName string) (string, []interface{})
}
```
其次用 map 保存不同数据库的方言，key 是数据库名称， value 是实现了接口的方言对象

```go
func RegisterDialect(name string, dialect Dialect)
```

最后提供获取方言的方法

```go
func GetDialect(name string) (dialect Dialect, ok bool)
```

### 3、实现对象和表的转换 (schema)
给定任意一个对象，转化为关系型数据库的表结构。对应关系如下：

|   映射对象   |   映射对象    |
|:--------:|:---------:|
|    表名    |   结构体名    |
| 字段名和字段类型 |  成员变量和类型  |
| 额外的约束条件  | 成员变量的 tag |

定义 Field 结构体来表示字段(列)的名称、类型、约束条件

```go
type Field struct {
    Name string //列字段名
    Type string // 列字段类型
    Tag  string // 列字段约束条件
}
```

定义 Schema 结构体来表示一张表的信息

```go
type Schema struct {
    Model      interface{}       // 被映射的对象
    Name       string            // 表名
    Fields     []*Field          // 多个列
    FieldNames []string          // 每个列的列名
    fieldMap   map[string]*Field // 存储列的信息，也就是 Field
}
```
实现核心的解析功能，将任意对象解析为一个 Schema 实例。
    
```go
func Parse(dest interface{}, d dialect.Dialect) *Schema
```

***首先***使用 reflect.ValueOf(dest) 拿到对象的值，再利用 reflect.Indirect 拿到指向这个值的指针，最后使用 .Type() 获取这个对象的类型，用 modelType 保存(包名.结构体名)(eg: schema.User)。modelType.Name() 就是这个对象的名称(eg: User)，也就是表名。

***然后*** modelType.NumField() 获取列的数量。之后利用 for range 遍历 modelType.Field(i) 拿到每一个列(Field)的信息， 并赋值给 Schema 的 Field，得到一个 Schema， 也就是一张表的在对应数据库的映射。

其中在获取列的信息的时候，需要用到之前定义的方言，将列的数据类型转化为数据库支持的数据类型。列字段的类型需要通过反射来创建：reflect.New(p.Type)

***最后*** 给 Session 添加 新增表 / 删除表的功能。

Model() 方法用于给 refTable 赋值。因为解析工作相对耗时，所以要将解析的结果保存在 refTable 中。在这个方法中调用了之前的解析方法。

```go
func (s *Session) Model(value interface{}) *Session
```

RefTable() 方法用于返回解析后的表，用于后续针对于表的操作。

```go
func (s *Session) RefTable() *schema.Schema
```

接下来实现数据库表的创建、删除和判断是否存在的功能。三个方法的实现逻辑是相似的，利用 RefTable() 返回的数据库表和字段的信息，拼接出 SQL 语句，调用原生 SQL 接口执行。

### 4、实现新增和查询
把新增和查询放在一个小节中是因为这两个功能的难点是相反的。
在实现新增和查询之前先实现一个子句的构造方式，因为新增或者查询语句一般是由很多子句构成的，例如

```sqlite
SELECT col1, col2, ... FROM table_name WHERE [ conditions ] GROUP BY col1 HAVING [ conditions ]
```

一次性的构造完整的 SQL 语句比较困难，可以先实现子句后再拼接成完整的 SQL 语句。

Set 方法根据传入的 type 去调用对应的 子句构造方法
    
```go
func (c *Clause) Set(name Type, vars ...interface{})
```

Build 方法根据传入的 type 的顺序来依次拼接成完整的 SQL 语句
    
```go
func (c *Clause) Build(orders ...Type) (string, []i nterface{})
```

新增的流程是先通过 RefTable() 拿到之前解析到的表，然后利用拿到的信息多次调用 clause.Set()构造出相应的子句，调用一次 clause.Build() 按照传入的顺序构造出最终的 SQL 语句，最后调用 Raw 格式化 SQL 语句，并调用Session 中的 Exec() 方法执行语句。

查询则反过来，使用反射(reflect)将数据库的记录转换为对应的结构体实例。具体如下：

    1) destSlice.Type().Elem() 获取切片的单个元素的类型 destType，使用 reflect.New() 方法创建一个 destType 的实例，作为 Model() 的入参，映射出表结构 RefTable()。
    
    2）根据表结构，使用 clause 构造出 SELECT 语句，查询到所有符合条件的记录 rows。
    
    3）遍历每一行记录，利用反射创建 destType 的实例 dest，将 dest 的所有字段平铺开，构造切片 values。
    
    4）调用 rows.Scan() 将该行记录每一列的值依次赋值给 values 中的每一个字段。
    
    5）将 dest 添加到切片 destSlice 中。循环直到所有的记录都添加到切片 destSlice 中。

新增的难点在于将已经存在的对象的每一个字段的值平铺开来。

查询的难点在于需要根据平铺开的字段的值构造出对象。

### 5、实现修改和删除以及统计功能
实现较为简单，类似于查询和新增方法。通过 RefTable() 获取表的信息，再利用这信息构造子句并拼接为完整的子句，最后用 Raw 方法将 SQL 子句和参数 格式化并执行。

注意链式调用。有一些语句例如 WHERE、LIMIT、ORDER BY 等比较适合链式调用，例如：

```sqlite
s.Where("Age > 18").Limit(3).Find(&users)
```

在这些子句的构造方法添加一个 Session 的返回值就可以实现链式调用了。
### 6、实现钩子功能
主要思想是提前在可能增加功能的地方埋好(预设)一个钩子，当我们需要重新修改或者增加这个地方的逻辑的时候，把扩展的类或者方法挂载到这个点即可。对于 ORM 框架来说，合适的扩展点在记录的增删查改前后。

钩子机制同样需要反射来实现。在 CallMethod() 中，当前会话正在操作的对象 s.RefTable().Model 使用 MethodByName(method) 方法利用反射得到该对象的钩子方法。这个参数 method 对应的是给对象编写的同名钩子方法，例如

```go
func (account *Account) AfterQuery(s *Session) error
```

然后利用 Call() 方法即可调用这个钩子方法，由于我们定义的钩子方法的参数是这个会话 Session，所以在调用 Call() 之前需要使用反射机制拿到 param，类型是 *Session。接下来，将 CallMethod() 方法在 Find、Insert、Update、Delete 方法内部调用即可。

```go
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
```

### 7、实现事务功能
数据库事务(transaction)是访问并可能操作各种数据项的一个数据库操作序列，这些操作要么全部执行,要么全部不执行，是一个不可分割的工作单位。事务由事务开始与事务结束之间执行的全部数据库操作组成。

一个数据库支持事务，那么必须具备 ACID 四个属性。

    1）原子性(Atomicity)：事务中的全部操作在数据库中是不可分割的，要么全部完成，要么全部不执行。
    2）一致性(Consistency): 几个并行执行的事务，其执行结果必须与按某一顺序 串行执行的结果相一致。
    3）隔离性(Isolation)：事务的执行不受其他事务的干扰，事务执行的中间结果对其他事务必须是透明的。
    4）持久性(Durability)：对于任意已提交事务，系统必须保证该事务对数据库的改变不被丢失，即使数据库出现故障。

BEGIN 开启事务，COMMIT 提交事务，ROLLBACK 回滚事务。

任何一个事务，均以 BEGIN 开始，COMMIT 或 ROLLBACK 结束。

之前直接使用 sql.DB 对象执行 SQL 语句，如果要支持事务，需要更改为 sql.Tx 执行。这里增加一个可选项，可以使用 Tx 执行，也可以使用 db执行。
需要定义一个接口，需要实现的方法是开始提到的三个原生方法。相当于实现了两种数据库的使用方式，但是这两种方式都需要这三个原生的方法。

接下来封装事务的 Begin、Commit 和 Rollback方法。

Begin 方法就是调用 s.db.Begin() 得到 *sql.Tx 对象。


## 三、实现 Engine 用于 Session 交互前的准备工作
Engine 是用户和 ORM 框架的交互入口，主要功能就是连接数据库并做一些交互前的准备工作，同时为用户提供一些功能的使用接口，Engine的结构如下：

```go
type Engine struct {
    db      *sql.DB         // 数据库句柄
    dialect dialect.Dialect // 添加 dialect 实现对不同数据库的支持
}
```
### 1、创建 Engine 用于准备工作或者全局的一些工作，包括连接数据库等

```go
func NewEngine(driver, source string) (e *Engine, err error) 
```

### 2、Engine 提供用于创建对应的 Session 的方法

```go
func (engine *Engine) NewSession() *session.Session
```

### 3、Engine 提供对封装的事务方法的调用，为用户提供一键式使用的接口

```go
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error)
```

### 4、Engine 提供对数据库迁移的调用，使用了之前的事务

```go
func (engine *Engine) Migrate(value interface{}) error
```

这里的数据库迁移只支持字段的新增和删除，不支持字段类型变更。在这个方法中计算新表和旧表的差集。新表有的旧表没有的就是新增字段，新表没有的旧表有的就是要删除的字段。找到要新增和删除的字段后就可以按照原生 SQL 的逻辑实现。

在原生 SQL 中 新增字段需要使用语句：

```sqlite
ALTER TABLE table_name ADD COLUMN col_name, col_type;
```
在原生 SQL 中 删除字段需要三步：
    
```go
// 1、从 old_table 中挑选需要保留的字段到 new_table 中
CREATE TABLE new_table AS SELECT col1, col2, ... from old_table
// 2、删除 old_table
DROP TABLE old_table
// 3、重命名 new_table 为 old_table
ALTER TABLE new_table RENAME TO old_table;
```

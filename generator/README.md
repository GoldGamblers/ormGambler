# generator包 说明
查询语句一般由很多个子句(clause) 构成，如果想一次构造出完整的 SQL 语句是比较困难的, 需要将构造 SQL 语句这一部分独立出来。
例如：

```sqlite
SELECT col1, col2, ...
    FROM table_name
    WHERE [ conditions ]
    GROUP BY col1
    HAVING [ conditions ]
```

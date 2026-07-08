# orm

一个轻量级的 MySQL ORM 框架（学习练习项目），基于 `database/sql` + `github.com/go-sql-driver/mysql` 实现，提供链式查询 API、struct tag 映射、事务、自动迁移与原生 SQL 兜底。

## 特性

- **链式查询**：`Where` / `OrWhere` / `Join` / `LeftJoin` / `RightJoin` / `Order` / `GroupBy` / `Having` / `Limit` / `Offset` / 行锁
- **struct tag 映射**：字段名、列名、主键、自增、nullable、omitempty
- **CRUD**：`Find` / `First` / `Insert` / `Update` / `Delete` / `Count`
- **事务**：`Transaction` 自动提交/回滚，`Begin` 手动控制
- **自动迁移**：`AutoMigrate` 自动建表、同步缺失列
- **原生 SQL**：`Raw` 兜底，支持 struct/slice 扫描
- **连接池**：可配置 `MaxOpenConns` / `MaxIdleConns` / 连接寿命
- **安全防护**：无 `where` 的 `Update`/`Delete` 直接报错，防止全表误操作

## 目录结构

| 文件 | 职责 |
|---|---|
| `errors.go` | 包内错误定义 |
| `model.go` | struct tag 反射解析、模型元数据缓存、驼峰转蛇形 |
| `db.go` | 连接管理、连接池、入口方法 |
| `query.go` | 链式查询构建器 + SQL 生成 |
| `executor.go` | 查询执行与结果扫描 |
| `tx.go` | 事务封装 |
| `raw.go` | 原生 SQL |
| `migrate.go` | 自动建表/同步列 |
| `orm_test.go` | 纯逻辑单元测试（无需 DB） |
| `integration_test.go` | 集成测试（`ORM_TEST_DSN` 控制） |

## 连接数据库

```go
import "goProject/orm"

// 简单方式
db, err := orm.Open("user:password@tcp(127.0.0.1:3306)/dbname?parseTime=true&charset=utf8mb4")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 带连接池配置
db, err = orm.OpenWithConfig(orm.Config{
    DSN:             "user:password@tcp(127.0.0.1:3306)/dbname?parseTime=true",
    MaxOpenConns:    100,
    MaxIdleConns:    20,
    ConnMaxLifetime: time.Hour,
    ConnMaxIdleTime: 30 * time.Minute,
})
```

> `Open` 默认会 `Ping` 一次数据库，连接失败立即返回错误。

## 模型定义

通过 struct tag 描述字段与数据库列的映射关系：

| tag 选项 | 说明 |
|---|---|
| `column:xxx` | 指定列名，省略则用字段名蛇形化 |
| `primaryKey` | 标记主键 |
| `autoIncrement` | 自增列，插入时零值会跳过 |
| `nullable` | 允许 NULL（文档说明，不影响 DDL 之外的逻辑） |
| `omitempty` | insert/update 时该字段为零值则跳过 |
| `-` | 忽略该字段 |

```go
type User struct {
    ID        int64     `orm:"column:id;primaryKey;autoIncrement"`
    Name      string    `orm:"column:name;omitempty"`
    Age       int       `orm:"column:age"`
    Email     string    `orm:"column:email;nullable"`
    CreatedAt time.Time `orm:"column:created_at"`
    IgnoreMe  string    `orm:"-"`
}
```

### 自定义表名

实现 `Tabler` 接口优先于默认蛇形规则：

```go
func (User) TableName() string { return "t_user" }
```

否则 `User` → `user`，`IntegrationUser` → `integration_user`。

## 查询

### 多条记录

```go
var users []User
err := db.Model(&User{}).
    Where("age > ?", 18).
    OrWhere("name LIKE ?", "%tom%").
    Order("id DESC").
    Limit(10).
    Offset(20).
    Find(ctx, &users)
```

也支持 `*[]*User`（元素为指针的切片）。

### 单条记录

```go
var u User
// 按主键
err := db.Model(&User{}).First(ctx, &u, 1)

// 按 where 条件
err = db.Model(&User{}).Where("name = ?", "tom").First(ctx, &u)

// 未找到时返回 orm.ErrNotFound
if errors.Is(err, orm.ErrNotFound) {
    // ...
}
```

### 聚合查询

```go
n, err := db.Model(&User{}).Where("age > ?", 18).Count(ctx)
```

### 链式方法一览

| 方法 | 说明 |
|---|---|
| `Select(cols...)` | 指定查询列 |
| `Where(expr, args...)` | AND 条件 |
| `OrWhere(expr, args...)` | OR 条件 |
| `Join(table, cond, args...)` | INNER JOIN |
| `LeftJoin` / `RightJoin` | 左/右连接 |
| `Order(col)` | 排序，支持 `"id DESC"` |
| `GroupBy(cols...)` | 分组 |
| `Having(expr, args...)` | HAVING |
| `Limit(n)` / `Offset(n)` | 分页 |
| `LockUpdate()` | `FOR UPDATE` 排他锁 |
| `LockShare()` | `FOR SHARE` 共享锁 |

## 写入

### 插入

支持 struct 或 `map[string]any`，返回自增 ID：

```go
// struct
id, err := db.Model(&User{}).Insert(ctx, &User{
    Name: "tom", Age: 20, CreatedAt: time.Now(),
})

// map
id, err = db.Table("t_user").Insert(ctx, map[string]any{
    "name": "jerry", "age": 25,
})
```

- 自增主键为零值时自动跳过，交给数据库生成
- `omitempty` 字段为零值时跳过

### 更新

```go
// map
rows, err := db.Model(&User{}).Where("id = ?", 1).
    Update(ctx, map[string]any{"age": 21})

// struct（主键默认不更新，omitempty 字段零值跳过）
rows, err = db.Model(&User{}).Where("id = ?", 1).
    Update(ctx, &User{Name: "newname"})
```

> **安全限制**：`Update` 必须带 `where`，否则返回 `orm.ErrMissingWhere`。如需全表更新，请显式 `Where("1=1")`。

### 删除

```go
rows, err := db.Model(&User{}).Where("id = ?", 1).Delete(ctx)
```

> 同样要求必须带 `where`。

## 事务

### 自动管理（推荐）

```go
err := db.Transaction(ctx, func(tx *orm.Tx) error {
    if _, err := tx.Model(&User{}).Insert(ctx, &User{Name: "a"}); err != nil {
        return err
    }
    if _, err := tx.Model(&Account{}).Where("id=?", 1).Update(ctx, map[string]any{"balance": 0}); err != nil {
        return err
    }
    return nil // 返回 nil 自动提交，返回 error 自动回滚
})
```

事务内的 `Tx` 与 `DB` 拥有相同的 `Model` / `Table` / `Raw` 方法，API 完全一致。

### 手动控制

```go
tx, err := db.Begin(ctx)
if err != nil { ... }
defer tx.Rollback() // 安全起见，Commit 成功后调用 Rollback 无副作用

// ... 使用 tx.Model(...) / tx.Table(...) ...

if err := tx.Commit(); err != nil { ... }
```

## 原生 SQL

当链式 API 不够用时，用 `Raw` 兜底：

```go
// 扫描到 struct
var u User
err := db.Raw("SELECT * FROM t_user WHERE name = ?", "tom").Scan(ctx, &u)

// 扫描到 slice
var users []User
err = db.Raw("SELECT * FROM t_user ORDER BY id").Scan(ctx, &users)

// 扫描到基本类型 slice
var ids []int64
err = db.Raw("SELECT id FROM t_user").Scan(ctx, &ids)

// 执行非查询
res, err := db.Raw("UPDATE t_user SET age = ? WHERE id = ?", 30, 1).Exec(ctx)
```

> `Raw().Scan()` 在无记录时返回 `sql.ErrNoRows`。

## 自动迁移

```go
err := db.AutoMigrate(ctx, &User{}, &Account{})
```

- 表不存在：根据模型生成 `CREATE TABLE`，含主键、类型映射、`ENGINE=InnoDB CHARSET=utf8mb4`
- 表已存在：对比 `information_schema.columns`，仅**新增**缺失的列（不删除、不改类型）

### Go 类型 → MySQL 类型映射

| Go 类型 | MySQL 类型 |
|---|---|
| `int64` / `int` | `BIGINT` |
| `int32` | `INT` |
| `int16` / `int8` | `SMALLINT` |
| `uint64` / `uint` | `BIGINT UNSIGNED` |
| `uint32` | `INT UNSIGNED` |
| `float32` | `FLOAT` |
| `float64` | `DOUBLE` |
| `bool` | `TINYINT(1)` |
| `string` | `VARCHAR(255)` |
| `time.Time` | `DATETIME` |

## 错误处理

```go
var (
    ErrNotFound        // First 未找到记录
    ErrMissingWhere    // Update/Delete 无 where
    ErrMissingTable    // 未指定表名
    ErrNotStruct       // 目标非 struct
    ErrNotSlice        // 目标非 slice
    ErrUnsupportedDest // Raw Scan 目标类型不支持
)
```

```go
if errors.Is(err, orm.ErrNotFound) {
    // 记录不存在
}
```

## 注意事项

1. **NULL 列**：可空列建议使用 `sql.NullString` / `sql.NullInt64` 或指针类型字段，否则扫描到 `string`/`int` 遇 NULL 会报错。
2. **map 顺序不定**：用 `map[string]any` 做插入/更新时，列顺序由 Go map 遍历决定，不保证稳定。如需稳定顺序，用 struct。
3. **连接复用**：`Query` 是可变对象，链式方法会修改自身。不要并发复用同一个 `Query`，每个查询新建。
4. **上下文**：所有终结方法（`Find`/`First`/`Insert`/`Update`/`Delete`/`Count`/`Scan`/`Exec`）都接收 `context.Context`，支持超时与取消。

## 运行测试

### 单元测试（无需数据库）

```cmd
go test ./orm/...
```

### 集成测试（需本地 MySQL）

```cmd
:: 创建测试数据库
:: mysql -uroot -p -e "CREATE DATABASE orm_test"

:: 设置连接串并运行
set ORM_TEST_DSN=root:password@tcp(127.0.0.1:3306)/orm_test?parseTime=true&charset=utf8mb4
go test ./orm/... -v -run Integration
```

集成测试覆盖：插入(struct/map) → 查询(Find/First/Count) → 更新 → 删除 → `ErrNotFound` → 事务回滚/提交 → `Raw` 扫描/执行。

## API 速查

```
orm.Open(dsn) (*DB, error)
orm.OpenWithConfig(Config) (*DB, error)

db.Table(name) *Query
db.Model(model) *Query
db.AutoMigrate(ctx, models...)
db.Transaction(ctx, func(*Tx) error) error
db.Begin(ctx) (*Tx, error)
db.Raw(sql, args...) *RawQuery
db.Close() / db.SqlDB() *sql.DB

q.Select().Where().OrWhere().Join().LeftJoin().RightJoin()
 .Order().GroupBy().Having().Limit().Offset().LockUpdate().LockShare()
q.Find(ctx, &slice)  error
q.First(ctx, &struct, pk?)  error
q.Count(ctx)  (int64, error)
q.Insert(ctx, data)  (int64, error)
q.Update(ctx, data)  (int64, error)
q.Delete(ctx)  (int64, error)

tx.Model(model) / tx.Table(name) / tx.Raw(sql)
tx.Commit() / tx.Rollback()

raw.Scan(ctx, dest)  error
raw.Exec(ctx)  (sql.Result, error)
```

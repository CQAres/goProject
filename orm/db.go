package orm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config 数据库连接配置
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DB 是 ORM 的入口，持有连接池
type DB struct {
	db     *sql.DB
	config Config
}

// Open 使用 DSN 打开数据库连接
func Open(dsn string) (*DB, error) {
	return OpenWithConfig(Config{DSN: dsn})
}

// OpenWithConfig 使用完整配置打开连接
func OpenWithConfig(c Config) (*DB, error) {
	if c.DSN == "" {
		return nil, fmt.Errorf("orm: empty DSN")
	}
	sqldb, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("orm: open mysql: %w", err)
	}
	applyPool(sqldb, c)
	if err := sqldb.Ping(); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("orm: ping mysql: %w", err)
	}
	return &DB{db: sqldb, config: c}, nil
}

func applyPool(db *sql.DB, c Config) {
	maxOpen := c.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 50
	}
	maxIdle := c.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 10
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	if c.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(c.ConnMaxLifetime)
	}
	if c.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	}
}

// Close 关闭连接池
func (db *DB) Close() error { return db.db.Close() }

// SqlDB 暴露底层 *sql.DB，便于使用原生 database/sql 能力
func (db *DB) SqlDB() *sql.DB { return db.db }

// Table 指定表名构建查询
func (db *DB) Table(name string) *Query {
	return &Query{db: db, table: name}
}

// Model 基于模型实例构建查询，自动解析表名与字段
func (db *DB) Model(model any) *Query {
	q := &Query{db: db}
	meta, err := parseModel(model)
	if err != nil {
		q.err = err
		return q
	}
	q.table = meta.TableName
	q.model = meta
	return q
}

// AutoMigrate 自动建表/同步列
func (db *DB) AutoMigrate(ctx context.Context, models ...any) error {
	return db.migrate(ctx, models)
}

// Transaction 执行事务，出错自动回滚
func (db *DB) Transaction(ctx context.Context, fn func(*Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

// Begin 开启事务
func (db *DB) Begin(ctx context.Context) (*Tx, error) {
	sqltx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{db: db, tx: sqltx}, nil
}

// Raw 原生 SQL 查询
func (db *DB) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{db: db, sqlStr: query, args: args}
}

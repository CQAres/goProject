package orm

import "database/sql"

// Tx 事务封装，复用 Query 的链式能力
type Tx struct {
	db *DB
	tx *sql.Tx
}

// Commit 提交事务
func (t *Tx) Commit() error { return t.tx.Commit() }

// Rollback 回滚事务
func (t *Tx) Rollback() error { return t.tx.Rollback() }

// Table 指定表名构建查询（在事务内执行）
func (t *Tx) Table(name string) *Query {
	return &Query{db: t.db, tx: t, table: name}
}

// Model 基于模型实例构建查询（在事务内执行）
func (t *Tx) Model(model any) *Query {
	q := &Query{db: t.db, tx: t}
	meta, err := parseModel(model)
	if err != nil {
		q.err = err
		return q
	}
	q.table = meta.TableName
	q.model = meta
	return q
}

// Raw 原生 SQL（在事务内执行）
func (t *Tx) Raw(query string, args ...any) *RawQuery {
	return &RawQuery{tx: t, sqlStr: query, args: args}
}

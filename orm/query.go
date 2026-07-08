package orm

import (
	"context"
	"database/sql"
	"strings"
)

// where 描述一个 WHERE 片段
type where struct {
	expr string
	args []any
	op   string // "AND" 或 "OR"
}

type join struct {
	typ   string // INNER / LEFT / RIGHT
	table string
	cond  string
	args  []any
}

type order struct {
	col  string
	desc bool
}

// execer 抽象 sql.DB 与 sql.Tx 共同具备的执行能力
type execer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Query 链式查询构建器
type Query struct {
	db    *DB
	tx    *Tx
	table string
	model *ModelMeta
	err   error // 解析阶段产生的错误

	selects []string
	wheres  []where
	joins   []join
	orders  []order
	groupBy []string
	having  []where
	limit   int
	offset  int
	lock    string
}

func (q *Query) execer() execer {
	if q.tx != nil {
		return q.tx.tx
	}
	return q.db.db
}

// ---- 链式方法（返回自身以支持链式调用）----

func (q *Query) Select(cols ...string) *Query {
	q.selects = append(q.selects, cols...)
	return q
}

func (q *Query) Where(expr string, args ...any) *Query {
	q.wheres = append(q.wheres, where{expr: expr, args: args, op: "AND"})
	return q
}

func (q *Query) OrWhere(expr string, args ...any) *Query {
	q.wheres = append(q.wheres, where{expr: expr, args: args, op: "OR"})
	return q
}

func (q *Query) Join(table, cond string, args ...any) *Query {
	q.joins = append(q.joins, join{typ: "INNER", table: table, cond: cond, args: args})
	return q
}

func (q *Query) LeftJoin(table, cond string, args ...any) *Query {
	q.joins = append(q.joins, join{typ: "LEFT", table: table, cond: cond, args: args})
	return q
}

func (q *Query) RightJoin(table, cond string, args ...any) *Query {
	q.joins = append(q.joins, join{typ: "RIGHT", table: table, cond: cond, args: args})
	return q
}

// Order 设置排序，支持 "id DESC" / "id ASC" / "id" 形式
func (q *Query) Order(col string) *Query {
	col = strings.TrimSpace(col)
	desc := false
	upper := strings.ToUpper(col)
	if strings.HasSuffix(upper, " DESC") {
		desc = true
		col = strings.TrimSpace(col[:len(col)-len(" DESC")])
	} else if strings.HasSuffix(upper, " ASC") {
		col = strings.TrimSpace(col[:len(col)-len(" ASC")])
	}
	q.orders = append(q.orders, order{col: col, desc: desc})
	return q
}

func (q *Query) GroupBy(cols ...string) *Query {
	q.groupBy = append(q.groupBy, cols...)
	return q
}

func (q *Query) Having(expr string, args ...any) *Query {
	q.having = append(q.having, where{expr: expr, args: args, op: "AND"})
	return q
}

func (q *Query) Limit(n int) *Query  { q.limit = n; return q }
func (q *Query) Offset(n int) *Query { q.offset = n; return q }

// LockUpdate 行级排他锁 (FOR UPDATE)
func (q *Query) LockUpdate() *Query { q.lock = "FOR UPDATE"; return q }

// LockShare 行级共享锁 (FOR SHARE)
func (q *Query) LockShare() *Query { q.lock = "FOR SHARE"; return q }

// ---- SQL 构建 ----

func (q *Query) selectColumns() []string {
	if len(q.selects) > 0 {
		return q.selects
	}
	if q.model != nil {
		return q.model.Columns
	}
	return []string{"*"}
}

func (q *Query) buildSelectSQL() (string, []any) {
	var b strings.Builder
	args := []any{}

	b.WriteString("SELECT ")
	b.WriteString(strings.Join(q.selectColumns(), ", "))
	b.WriteString(" FROM ")
	b.WriteString(quoteIdent(q.table))

	for _, j := range q.joins {
		b.WriteByte(' ')
		b.WriteString(j.typ)
		b.WriteString(" JOIN ")
		b.WriteString(quoteIdent(j.table))
		b.WriteString(" ON ")
		b.WriteString(j.cond)
		args = append(args, j.args...)
	}

	wsql, wargs := q.buildWhere()
	if wsql != "" {
		b.WriteString(" WHERE ")
		b.WriteString(wsql)
		args = append(args, wargs...)
	}

	if len(q.groupBy) > 0 {
		b.WriteString(" GROUP BY ")
		b.WriteString(strings.Join(q.groupBy, ", "))
	}
	if len(q.having) > 0 {
		b.WriteString(" HAVING ")
		hsql, hargs := composeWheres(q.having)
		b.WriteString(hsql)
		args = append(args, hargs...)
	}
	if len(q.orders) > 0 {
		b.WriteString(" ORDER BY ")
		for i, o := range q.orders {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(o.col)
			if o.desc {
				b.WriteString(" DESC")
			}
		}
	}
	if q.limit > 0 {
		b.WriteString(" LIMIT ?")
		args = append(args, q.limit)
	}
	if q.offset > 0 {
		if q.limit == 0 {
			// MySQL 要求 OFFSET 必须搭配 LIMIT
			b.WriteString(" LIMIT 18446744073709551615")
		}
		b.WriteString(" OFFSET ?")
		args = append(args, q.offset)
	}
	if q.lock != "" {
		b.WriteByte(' ')
		b.WriteString(q.lock)
	}
	return b.String(), args
}

func (q *Query) buildWhere() (string, []any) {
	if len(q.wheres) == 0 {
		return "", nil
	}
	return composeWheres(q.wheres)
}

func composeWheres(ws []where) (string, []any) {
	var b strings.Builder
	var args []any
	for i, w := range ws {
		if i > 0 {
			b.WriteByte(' ')
			b.WriteString(w.op)
			b.WriteByte(' ')
		}
		b.WriteByte('(')
		b.WriteString(w.expr)
		b.WriteByte(')')
		args = append(args, w.args...)
	}
	return b.String(), args
}

func (q *Query) buildCountSQL() (string, []any) {
	var b strings.Builder
	args := []any{}
	b.WriteString("SELECT COUNT(*) FROM ")
	b.WriteString(quoteIdent(q.table))
	for _, j := range q.joins {
		b.WriteByte(' ')
		b.WriteString(j.typ)
		b.WriteString(" JOIN ")
		b.WriteString(quoteIdent(j.table))
		b.WriteString(" ON ")
		b.WriteString(j.cond)
		args = append(args, j.args...)
	}
	wsql, wargs := q.buildWhere()
	if wsql != "" {
		b.WriteString(" WHERE ")
		b.WriteString(wsql)
		args = append(args, wargs...)
	}
	return b.String(), args
}

func (q *Query) buildDeleteSQL() (string, []any) {
	var b strings.Builder
	b.WriteString("DELETE FROM ")
	b.WriteString(quoteIdent(q.table))
	wsql, wargs := q.buildWhere()
	if wsql != "" {
		b.WriteString(" WHERE ")
		b.WriteString(wsql)
	}
	return b.String(), wargs
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	return "?" + strings.Repeat(", ?", n-1)
}

// quoteIdent 为标识符加反引号；已是表达式(含括号/空格/逗号)的不加
func quoteIdent(name string) string {
	if name == "" {
		return name
	}
	if strings.ContainsAny(name, " ()`,") {
		return name
	}
	return "`" + name + "`"
}

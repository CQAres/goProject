package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// Find 将多条记录扫描到 dest（*[]Struct 或 *[]*Struct）
func (q *Query) Find(ctx context.Context, dest any) error {
	if q.err != nil {
		return q.err
	}
	sliceVal := reflect.ValueOf(dest)
	if sliceVal.Kind() != reflect.Ptr || sliceVal.IsNil() {
		return ErrNotSlice
	}
	sliceVal = sliceVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return ErrNotSlice
	}
	elemTyp := sliceVal.Type().Elem()
	isPtr := elemTyp.Kind() == reflect.Ptr
	baseTyp := elemTyp
	if isPtr {
		baseTyp = baseTyp.Elem()
	}
	if baseTyp.Kind() != reflect.Struct {
		return ErrNotStruct
	}
	meta, err := parseModel(reflect.New(baseTyp).Interface())
	if err != nil {
		return err
	}
	if q.model == nil {
		q.model = meta
	}
	if q.table == "" {
		q.table = meta.TableName
	}

	queryStr, args := q.buildSelectSQL()
	rows, err := q.execer().QueryContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("orm: query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		elem := reflect.New(baseTyp)
		if err := scanRows(rows, elem, meta); err != nil {
			return err
		}
		if isPtr {
			sliceVal.Set(reflect.Append(sliceVal, elem))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, elem.Elem()))
		}
	}
	return rows.Err()
}

// First 取第一条记录到 dest（*Struct）。可选传入主键值。
func (q *Query) First(ctx context.Context, dest any, pk ...any) error {
	if q.err != nil {
		return q.err
	}
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return ErrNotStruct
	}
	meta, err := parseModel(dest)
	if err != nil {
		return err
	}
	if q.model == nil {
		q.model = meta
	}
	if q.table == "" {
		q.table = meta.TableName
	}
	if len(pk) > 0 && meta.Primary != nil {
		q.Where(meta.Primary.Column+" = ?", pk[0])
	}
	q.limit = 1

	// 使用 Query + Next 而非 QueryRow，以便获取列名做映射
	queryStr, args := q.buildSelectSQL()
	rows, err := q.execer().QueryContext(ctx, queryStr, args...)
	if err != nil {
		return fmt.Errorf("orm: query: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return ErrNotFound
	}
	if err := scanRows(rows, destVal, meta); err != nil {
		return err
	}
	return rows.Err()
}

// Count 统计符合条件的记录数
func (q *Query) Count(ctx context.Context) (int64, error) {
	if q.err != nil {
		return 0, q.err
	}
	queryStr, args := q.buildCountSQL()
	var count int64
	if err := q.execer().QueryRowContext(ctx, queryStr, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("orm: count: %w", err)
	}
	return count, nil
}

// Insert 插入记录，data 可为 struct 或 map[string]any。返回自增ID。
func (q *Query) Insert(ctx context.Context, data any) (int64, error) {
	if q.err != nil {
		return 0, q.err
	}
	if q.table == "" {
		return 0, ErrMissingTable
	}
	cols, vals := buildInsertData(data)
	if len(cols) == 0 {
		return 0, fmt.Errorf("orm: nothing to insert")
	}
	var b strings.Builder
	b.WriteString("INSERT INTO ")
	b.WriteString(quoteIdent(q.table))
	b.WriteString(" (")
	b.WriteString(joinQuoted(cols))
	b.WriteString(") VALUES (")
	b.WriteString(placeholders(len(cols)))
	b.WriteString(")")

	res, err := q.execer().ExecContext(ctx, b.String(), vals...)
	if err != nil {
		return 0, fmt.Errorf("orm: insert: %w", err)
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// Update 更新记录，data 可为 struct 或 map[string]any。返回受影响行数。
func (q *Query) Update(ctx context.Context, data any) (int64, error) {
	if q.err != nil {
		return 0, q.err
	}
	if q.table == "" {
		return 0, ErrMissingTable
	}
	cols, vals := buildUpdateData(data)
	if len(cols) == 0 {
		return 0, fmt.Errorf("orm: nothing to update")
	}
	if len(q.wheres) == 0 {
		return 0, ErrMissingWhere
	}
	args := append([]any{}, vals...)
	var b strings.Builder
	b.WriteString("UPDATE ")
	b.WriteString(quoteIdent(q.table))
	b.WriteString(" SET ")
	for i, c := range cols {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(quoteIdent(c))
		b.WriteString(" = ?")
	}
	wsql, wargs := q.buildWhere()
	b.WriteString(" WHERE ")
	b.WriteString(wsql)
	args = append(args, wargs...)

	res, err := q.execer().ExecContext(ctx, b.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("orm: update: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// Delete 删除记录。返回受影响行数。
func (q *Query) Delete(ctx context.Context) (int64, error) {
	if q.err != nil {
		return 0, q.err
	}
	if q.table == "" {
		return 0, ErrMissingTable
	}
	if len(q.wheres) == 0 {
		return 0, ErrMissingWhere
	}
	queryStr, args := q.buildDeleteSQL()
	res, err := q.execer().ExecContext(ctx, queryStr, args...)
	if err != nil {
		return 0, fmt.Errorf("orm: delete: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// buildInsertData 从 struct/map 提取需要插入的列与值
func buildInsertData(data any) ([]string, []any) {
	if m, ok := data.(map[string]any); ok {
		cols := make([]string, 0, len(m))
		vals := make([]any, 0, len(m))
		for k, val := range m {
			cols = append(cols, k)
			vals = append(vals, val)
		}
		return cols, vals
	}

	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil
	}
	meta, err := parseModel(data)
	if err != nil {
		return nil, nil
	}
	cols := make([]string, 0, len(meta.Fields))
	vals := make([]any, 0, len(meta.Fields))
	for _, f := range meta.Fields {
		fv := rv.FieldByIndex(f.FieldIndex)
		// 自增字段如果是零值则跳过，交给数据库生成
		if f.IsAutoIncr && isZero(fv) {
			continue
		}
		if f.OmitEmpty && isZero(fv) {
			continue
		}
		cols = append(cols, f.Column)
		vals = append(vals, fv.Interface())
	}
	return cols, vals
}

// buildUpdateData 从 struct/map 提取需要更新的列与值
func buildUpdateData(data any) ([]string, []any) {
	if m, ok := data.(map[string]any); ok {
		cols := make([]string, 0, len(m))
		vals := make([]any, 0, len(m))
		for k, val := range m {
			cols = append(cols, k)
			vals = append(vals, val)
		}
		return cols, vals
	}

	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil
	}
	meta, err := parseModel(data)
	if err != nil {
		return nil, nil
	}
	cols := make([]string, 0, len(meta.Fields))
	vals := make([]any, 0, len(meta.Fields))
	for _, f := range meta.Fields {
		// 主键通常不更新
		if f.IsPrimary {
			continue
		}
		fv := rv.FieldByIndex(f.FieldIndex)
		if f.OmitEmpty && isZero(fv) {
			continue
		}
		cols = append(cols, f.Column)
		vals = append(vals, fv.Interface())
	}
	return cols, vals
}

func isZero(v reflect.Value) bool {
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		return v.IsNil()
	}
	return v.IsZero()
}

func joinQuoted(cols []string) string {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = quoteIdent(c)
	}
	return strings.Join(parts, ", ")
}

// scanRows 扫描一行到 v（指向 struct 的 reflect.Value）
func scanRows(rows *sql.Rows, v reflect.Value, meta *ModelMeta) error {
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	scanArgs := make([]any, len(cols))
	for i, c := range cols {
		fm := meta.FieldByColumn(c)
		if fm == nil {
			// 模型中没有该列，用 dummy 接住
			var dummy any
			scanArgs[i] = &dummy
			continue
		}
		fv := v.Elem().FieldByIndex(fm.FieldIndex)
		scanArgs[i] = fv.Addr().Interface()
	}
	return rows.Scan(scanArgs...)
}

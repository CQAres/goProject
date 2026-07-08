package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
)

// RawQuery 原生 SQL 查询
type RawQuery struct {
	db     *DB
	tx     *Tx
	sqlStr string
	args   []any
}

func (r *RawQuery) execer() execer {
	if r.tx != nil {
		return r.tx.tx
	}
	return r.db.db
}

// Scan 将查询结果扫描到 dest（struct、*[]Struct、*[]*Struct 或基本类型切片）
func (r *RawQuery) Scan(ctx context.Context, dest any) error {
	rv := reflect.ValueOf(dest)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrUnsupportedDest
	}
	elem := rv.Elem()
	if elem.Kind() == reflect.Slice {
		return r.scanSlice(ctx, elem)
	}
	if elem.Kind() == reflect.Struct {
		return r.scanStruct(ctx, rv)
	}
	return ErrUnsupportedDest
}

// Exec 执行非查询 SQL（INSERT/UPDATE/DELETE/DDL），返回 sql.Result
func (r *RawQuery) Exec(ctx context.Context) (sql.Result, error) {
	res, err := r.execer().ExecContext(ctx, r.sqlStr, r.args...)
	if err != nil {
		return nil, fmt.Errorf("orm: raw exec: %w", err)
	}
	return res, nil
}

func (r *RawQuery) scanSlice(ctx context.Context, elem reflect.Value) error {
	elemTyp := elem.Type().Elem()
	isPtr := elemTyp.Kind() == reflect.Ptr
	baseTyp := elemTyp
	if isPtr {
		baseTyp = baseTyp.Elem()
	}
	if baseTyp.Kind() != reflect.Struct {
		return r.scanPrimitiveSlice(ctx, elem, baseTyp)
	}
	meta, err := parseModel(reflect.New(baseTyp).Interface())
	if err != nil {
		return err
	}
	rows, err := r.execer().QueryContext(ctx, r.sqlStr, r.args...)
	if err != nil {
		return fmt.Errorf("orm: raw query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		e := reflect.New(baseTyp)
		if err := scanRows(rows, e, meta); err != nil {
			return err
		}
		if isPtr {
			elem.Set(reflect.Append(elem, e))
		} else {
			elem.Set(reflect.Append(elem, e.Elem()))
		}
	}
	return rows.Err()
}

func (r *RawQuery) scanPrimitiveSlice(ctx context.Context, elem reflect.Value, baseTyp reflect.Type) error {
	rows, err := r.execer().QueryContext(ctx, r.sqlStr, r.args...)
	if err != nil {
		return fmt.Errorf("orm: raw query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		v := reflect.New(baseTyp)
		if err := rows.Scan(v.Interface()); err != nil {
			return err
		}
		elem.Set(reflect.Append(elem, v.Elem()))
	}
	return rows.Err()
}

func (r *RawQuery) scanStruct(ctx context.Context, rv reflect.Value) error {
	meta, err := parseModel(rv.Interface())
	if err != nil {
		return err
	}
	rows, err := r.execer().QueryContext(ctx, r.sqlStr, r.args...)
	if err != nil {
		return fmt.Errorf("orm: raw query: %w", err)
	}
	defer rows.Close()
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	return scanRows(rows, rv, meta)
}

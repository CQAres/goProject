package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

func (db *DB) migrate(ctx context.Context, models []any) error {
	for _, m := range models {
		meta, err := parseModel(m)
		if err != nil {
			return err
		}
		exists, err := db.tableExists(ctx, meta.TableName)
		if err != nil {
			return err
		}
		if !exists {
			if err := db.createTable(ctx, meta); err != nil {
				return err
			}
			continue
		}
		if err := db.syncColumns(ctx, meta); err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) tableExists(ctx context.Context, name string) (bool, error) {
	var n int
	err := db.db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
		name,
	).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("orm: check table exists: %w", err)
	}
	return n > 0, nil
}

func (db *DB) createTable(ctx context.Context, meta *ModelMeta) error {
	var b strings.Builder
	fmt.Fprintf(&b, "CREATE TABLE `%s` (", meta.TableName)
	for i, f := range meta.Fields {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "`%s` %s", f.Column, mysqlType(f))
		// 自增字段必须 NOT NULL
		if f.IsAutoIncr || !f.Nullable {
			b.WriteString(" NOT NULL")
		}
		if f.IsAutoIncr {
			b.WriteString(" AUTO_INCREMENT")
		}
	}
	if meta.Primary != nil {
		fmt.Fprintf(&b, ", PRIMARY KEY (`%s`)", meta.Primary.Column)
	}
	b.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4")
	if _, err := db.db.ExecContext(ctx, b.String()); err != nil {
		return fmt.Errorf("orm: create table %s: %w", meta.TableName, err)
	}
	return nil
}

// syncColumns 同步缺失的列（仅新增，不删除/不修改类型）
func (db *DB) syncColumns(ctx context.Context, meta *ModelMeta) error {
	rows, err := db.db.QueryContext(ctx,
		"SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ?",
		meta.TableName,
	)
	if err != nil {
		return fmt.Errorf("orm: load columns: %w", err)
	}
	defer rows.Close()
	existing := map[string]bool{}
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return err
		}
		existing[col] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, f := range meta.Fields {
		if existing[f.Column] {
			continue
		}
		ddl := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", meta.TableName, f.Column, mysqlType(f))
		if !f.Nullable {
			ddl += " NOT NULL"
		}
		if _, err := db.db.ExecContext(ctx, ddl); err != nil {
			return fmt.Errorf("orm: add column %s.%s: %w", meta.TableName, f.Column, err)
		}
	}
	return nil
}

// mysqlType 将 Go 类型映射到 MySQL 类型
func mysqlType(f *FieldMeta) string {
	if f.Type == timeType {
		return "DATETIME"
	}
	switch f.Type.Kind() {
	case reflect.Int, reflect.Int64:
		return "BIGINT"
	case reflect.Int32:
		return "INT"
	case reflect.Int16, reflect.Int8:
		return "SMALLINT"
	case reflect.Uint, reflect.Uint64:
		return "BIGINT UNSIGNED"
	case reflect.Uint32:
		return "INT UNSIGNED"
	case reflect.Uint16, reflect.Uint8:
		return "SMALLINT UNSIGNED"
	case reflect.Float32:
		return "FLOAT"
	case reflect.Float64:
		return "DOUBLE"
	case reflect.Bool:
		return "TINYINT(1)"
	case reflect.String:
		return "VARCHAR(255)"
	case reflect.Ptr, reflect.Slice, reflect.Map:
		return "VARCHAR(255)"
	default:
		return "TEXT"
	}
}

package orm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// 集成测试模型
type IntegrationUser struct {
	ID        int64     `orm:"column:id;primaryKey;autoIncrement"`
	Name      string    `orm:"column:name;omitempty"`
	Age       int       `orm:"column:age"`
	Email     string    `orm:"column:email;nullable"`
	CreatedAt time.Time `orm:"column:created_at"`
}

func (IntegrationUser) TableName() string { return "orm_integration_users" }

// dsnFromEnv 从环境变量读取 DSN，未设置则跳过
func dsnFromEnv(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("ORM_TEST_DSN")
	if dsn == "" {
		t.Skip("ORM_TEST_DSN not set; skipping integration test")
	}
	return dsn
}

// openTestDB 打开连接、自动迁移、清理表
func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(dsnFromEnv(t))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	ctx := context.Background()
	if err := db.AutoMigrate(ctx, &IntegrationUser{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := db.SqlDB().ExecContext(ctx, "TRUNCATE TABLE orm_integration_users"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.SqlDB().ExecContext(context.Background(), "DROP TABLE IF EXISTS orm_integration_users")
		db.Close()
	})
	return db
}

func TestIntegration_InsertFindFirstUpdateDelete(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	// Insert (struct)
	id, err := db.Model(&IntegrationUser{}).Insert(ctx, &IntegrationUser{
		Name: "tom", Age: 20, Email: "tom@x.com", CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("insert struct: %v", err)
	}
	if id <= 0 {
		t.Fatalf("unexpected id %d", id)
	}

	// Insert (map)
	_, err = db.Table("orm_integration_users").Insert(ctx, map[string]any{
		"name": "jerry", "age": 25, "email": "jerry@x.com", "created_at": time.Now(),
	})
	if err != nil {
		t.Fatalf("insert map: %v", err)
	}

	// Find
	var users []IntegrationUser
	if err := db.Model(&IntegrationUser{}).Where("age >= ?", 20).Order("id ASC").Find(ctx, &users); err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	// First by PK
	var u IntegrationUser
	if err := db.Model(&IntegrationUser{}).First(ctx, &u, id); err != nil {
		t.Fatalf("first by pk: %v", err)
	}
	if u.Name != "tom" {
		t.Fatalf("name = %q", u.Name)
	}

	// First with where
	var u2 IntegrationUser
	if err := db.Model(&IntegrationUser{}).Where("name = ?", "jerry").First(ctx, &u2); err != nil {
		t.Fatalf("first by where: %v", err)
	}
	if u2.Age != 25 {
		t.Fatalf("age = %d", u2.Age)
	}

	// Count
	n, err := db.Model(&IntegrationUser{}).Count(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}

	// Update
	rows, err := db.Model(&IntegrationUser{}).Where("id = ?", id).Update(ctx, map[string]any{"age": 21})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if rows != 1 {
		t.Fatalf("update rows = %d, want 1", rows)
	}
	var u3 IntegrationUser
	if err := db.Model(&IntegrationUser{}).First(ctx, &u3, id); err != nil {
		t.Fatalf("first after update: %v", err)
	}
	if u3.Age != 21 {
		t.Fatalf("age = %d, want 21", u3.Age)
	}

	// Delete
	rows, err = db.Model(&IntegrationUser{}).Where("id = ?", id).Delete(ctx)
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if rows != 1 {
		t.Fatalf("delete rows = %d, want 1", rows)
	}
	n, _ = db.Model(&IntegrationUser{}).Count(ctx)
	if n != 1 {
		t.Fatalf("count after delete = %d, want 1", n)
	}

	// NotFound
	var u4 IntegrationUser
	if err := db.Model(&IntegrationUser{}).Where("name = ?", "nobody").First(ctx, &u4); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestIntegration_Transaction(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	_, _ = db.Model(&IntegrationUser{}).Insert(ctx, &IntegrationUser{Name: "tx", Age: 1, CreatedAt: time.Now()})

	// 失败事务应回滚
	err := db.Transaction(ctx, func(tx *Tx) error {
		if _, err := tx.Model(&IntegrationUser{}).Where("name = ?", "tx").Update(ctx, map[string]any{"age": 99}); err != nil {
			return err
		}
		return fmt.Errorf("intentional rollback")
	})
	if err == nil {
		t.Fatal("expected error from transaction")
	}
	var u IntegrationUser
	if err := db.Model(&IntegrationUser{}).Where("name = ?", "tx").First(ctx, &u); err != nil {
		t.Fatalf("first: %v", err)
	}
	if u.Age != 1 {
		t.Fatalf("age should be 1 after rollback, got %d", u.Age)
	}

	// 成功事务
	err = db.Transaction(ctx, func(tx *Tx) error {
		_, err := tx.Table("orm_integration_users").Insert(ctx, map[string]any{
			"name": "tx2", "age": 2, "created_at": time.Now(),
		})
		return err
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	n, _ := db.Model(&IntegrationUser{}).Count(ctx)
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
}

func TestIntegration_Raw(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	_, _ = db.Model(&IntegrationUser{}).Insert(ctx, &IntegrationUser{Name: "raw", Age: 7, CreatedAt: time.Now()})

	// Raw scan struct
	var u IntegrationUser
	if err := db.Raw("SELECT * FROM orm_integration_users WHERE name = ?", "raw").Scan(ctx, &u); err != nil {
		t.Fatalf("raw scan struct: %v", err)
	}
	if u.Age != 7 {
		t.Fatalf("age = %d, want 7", u.Age)
	}

	// Raw scan slice
	var users []IntegrationUser
	if err := db.Raw("SELECT * FROM orm_integration_users ORDER BY id").Scan(ctx, &users); err != nil {
		t.Fatalf("raw scan slice: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("len = %d, want 1", len(users))
	}

	// Raw exec
	res, err := db.Raw("UPDATE orm_integration_users SET age = ? WHERE name = ?", 8, "raw").Exec(ctx)
	if err != nil {
		t.Fatalf("raw exec: %v", err)
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		t.Fatalf("affected = %d, want 1", affected)
	}
}

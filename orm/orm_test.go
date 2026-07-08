package orm

import (
	"reflect"
	"testing"
)

// ---- 测试用模型 ----

type testUser struct {
	ID       int64  `orm:"column:id;primaryKey;autoIncrement"`
	Name     string `orm:"column:name;omitempty"`
	Age      int    `orm:"column:age"`
	Email    string `orm:"column:email;nullable"`
	IgnoreMe string `orm:"-"`
}

type withTabler struct {
	ID int64 `orm:"column:id;primaryKey"`
}

func (withTabler) TableName() string { return "custom_table" }

// ---- toSnake ----

func TestToSnake(t *testing.T) {
	cases := map[string]string{
		"":           "",
		"ID":         "id",
		"UserID":     "user_id",
		"Name":       "name",
		"HTTPServer": "http_server",
		"userName":   "user_name",
		"A":          "a",
	}
	for in, want := range cases {
		if got := toSnake(in); got != want {
			t.Errorf("toSnake(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseTagOpts(t *testing.T) {
	opts := parseTagOpts("column:name;primaryKey;omitempty;size:64")
	if opts["column"] != "name" || opts["size"] != "64" {
		t.Errorf("opts = %v", opts)
	}
	if _, ok := opts["primaryKey"]; !ok {
		t.Error("primaryKey missing")
	}
}

// ---- parseModel ----

func TestParseModel(t *testing.T) {
	meta, err := parseModel(testUser{})
	if err != nil {
		t.Fatal(err)
	}
	if meta.TableName != "test_user" {
		t.Errorf("table = %q, want test_user", meta.TableName)
	}
	if meta.Primary == nil || meta.Primary.Column != "id" || !meta.Primary.IsAutoIncr {
		t.Errorf("primary wrong: %+v", meta.Primary)
	}
	if m := meta.FieldByColumn("name"); m == nil || !m.OmitEmpty {
		t.Errorf("name field wrong: %+v", m)
	}
	if m := meta.FieldByColumn("email"); m == nil || !m.Nullable {
		t.Errorf("email field wrong: %+v", m)
	}
	if m := meta.FieldByColumn("ignore_me"); m != nil {
		t.Errorf("ignored field should be absent")
	}
	if len(meta.Fields) != 4 {
		t.Errorf("fields count = %d, want 4", len(meta.Fields))
	}
}

func TestParseModel_Tabler(t *testing.T) {
	meta, err := parseModel(withTabler{})
	if err != nil {
		t.Fatal(err)
	}
	if meta.TableName != "custom_table" {
		t.Errorf("table = %q, want custom_table", meta.TableName)
	}
}

func TestParseModel_Cache(t *testing.T) {
	m1, _ := parseModel(testUser{})
	m2, _ := parseModel(&testUser{})
	if m1 != m2 {
		t.Error("model cache should return same pointer")
	}
}

// ---- SQL 构建（无需 DB 连接）----

func TestBuildSelectSQL(t *testing.T) {
	q := (&DB{}).Table("users").
		Select("id", "name").
		Where("age > ?", 18).
		OrWhere("name LIKE ?", "%tom%").
		GroupBy("age").
		Having("COUNT(*) > ?", 1).
		Order("id DESC").
		Limit(10).
		Offset(5)

	sqlStr, args := q.buildSelectSQL()
	wantSQL := "SELECT id, name FROM `users` WHERE (age > ?) OR (name LIKE ?) GROUP BY age HAVING (COUNT(*) > ?) ORDER BY id DESC LIMIT ? OFFSET ?"
	if sqlStr != wantSQL {
		t.Errorf("sql mismatch:\n got: %s\nwant: %s", sqlStr, wantSQL)
	}
	wantArgs := []any{18, "%tom%", 1, 10, 5}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %v, want %v", args, wantArgs)
	}
}

func TestBuildSelectSQL_Join(t *testing.T) {
	q := (&DB{}).Table("orders").
		Select("orders.id", "users.name").
		LeftJoin("users", "users.id = orders.user_id").
		Where("orders.amount > ?", 100)
	sqlStr, args := q.buildSelectSQL()
	wantSQL := "SELECT orders.id, users.name FROM `orders` LEFT JOIN `users` ON users.id = orders.user_id WHERE (orders.amount > ?)"
	if sqlStr != wantSQL {
		t.Errorf("sql mismatch:\n got: %s\nwant: %s", sqlStr, wantSQL)
	}
	if !reflect.DeepEqual(args, []any{100}) {
		t.Errorf("args = %v", args)
	}
}

func TestBuildCountSQL(t *testing.T) {
	q := (&DB{}).Table("users").Where("age > ?", 18)
	sqlStr, args := q.buildCountSQL()
	want := "SELECT COUNT(*) FROM `users` WHERE (age > ?)"
	if sqlStr != want {
		t.Errorf("got: %s, want: %s", sqlStr, want)
	}
	if !reflect.DeepEqual(args, []any{18}) {
		t.Errorf("args = %v", args)
	}
}

func TestBuildDeleteSQL(t *testing.T) {
	q := (&DB{}).Table("users").Where("id = ?", 1)
	sqlStr, args := q.buildDeleteSQL()
	want := "DELETE FROM `users` WHERE (id = ?)"
	if sqlStr != want {
		t.Errorf("got: %s, want: %s", sqlStr, want)
	}
	if !reflect.DeepEqual(args, []any{1}) {
		t.Errorf("args = %v", args)
	}
}

func TestBuildInsertData_Struct(t *testing.T) {
	cols, vals := buildInsertData(&testUser{ID: 0, Name: "tom", Age: 20, Email: ""})
	// ID 自增且零值跳过; Name omitempty 但非空保留; Age 保留; Email 非 omitempty 零值保留
	wantCols := []string{"name", "age", "email"}
	if !reflect.DeepEqual(cols, wantCols) {
		t.Errorf("cols = %v, want %v", cols, wantCols)
	}
	if len(vals) != 3 || vals[0] != "tom" || vals[1] != 20 || vals[2] != "" {
		t.Errorf("vals = %v", vals)
	}
}

func TestBuildInsertData_Map(t *testing.T) {
	cols, vals := buildInsertData(map[string]any{"name": "tom", "age": 20})
	if len(cols) != 2 || len(vals) != 2 {
		t.Fatalf("cols=%v vals=%v", cols, vals)
	}
}

func TestBuildUpdateData_Struct(t *testing.T) {
	cols, vals := buildUpdateData(&testUser{ID: 1, Name: "tom", Age: 20, Email: ""})
	// 主键 id 跳过; email 非 omitempty 零值保留
	wantCols := []string{"name", "age", "email"}
	if !reflect.DeepEqual(cols, wantCols) {
		t.Errorf("cols = %v, want %v", cols, wantCols)
	}
	if len(vals) != 3 {
		t.Errorf("vals = %v", vals)
	}
}

func TestQuoteIdent(t *testing.T) {
	if got := quoteIdent("users"); got != "`users`" {
		t.Errorf("got %q", got)
	}
	if got := quoteIdent("COUNT(*)"); got != "COUNT(*)" {
		t.Errorf("got %q", got)
	}
	if got := quoteIdent("a, b"); got != "a, b" {
		t.Errorf("got %q", got)
	}
}

func TestPlaceholders(t *testing.T) {
	if placeholders(0) != "" {
		t.Error("placeholder 0 should be empty")
	}
	if placeholders(1) != "?" {
		t.Errorf("placeholder 1 = %q", placeholders(1))
	}
	if placeholders(3) != "?, ?, ?" {
		t.Errorf("placeholder 3 = %q", placeholders(3))
	}
}

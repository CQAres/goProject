package orm

import (
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

// FieldMeta 描述结构体单个字段与数据库列的映射关系
type FieldMeta struct {
	GoName     string       // Go 字段名
	Column     string       // 数据库列名
	Type       reflect.Type // Go 类型
	IsPrimary  bool         // 是否主键
	IsAutoIncr bool         // 是否自增
	Nullable   bool         // 允许 NULL
	OmitEmpty  bool         // 零值时跳过 insert/update
	FieldIndex []int        // reflect 字段索引
}

// ModelMeta 描述一个模型的整体映射信息
type ModelMeta struct {
	TableName string
	StructTyp reflect.Type
	Fields    []*FieldMeta
	Columns   []string
	Primary   *FieldMeta
}

// FieldByColumn 按列名查找字段
func (m *ModelMeta) FieldByColumn(col string) *FieldMeta {
	for _, f := range m.Fields {
		if f.Column == col {
			return f
		}
	}
	return nil
}

// Tabler 可选接口：模型可实现 TableName() 自定义表名
type Tabler interface {
	TableName() string
}

var (
	modelCache sync.Map // reflect.Type -> *ModelMeta
	timeType   = reflect.TypeOf(time.Time{})
)

// parseModel 解析模型的元数据，结果缓存
func parseModel(model any) (*ModelMeta, error) {
	typ, err := derefStructType(model)
	if err != nil {
		return nil, err
	}
	if cached, ok := modelCache.Load(typ); ok {
		return cached.(*ModelMeta), nil
	}
	return buildModelMeta(typ)
}

func derefStructType(v any) (reflect.Type, error) {
	if v == nil {
		return nil, ErrNotStruct
	}
	typ := reflect.TypeOf(v)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}
	return typ, nil
}

func buildModelMeta(typ reflect.Type) (*ModelMeta, error) {
	meta := &ModelMeta{
		StructTyp: typ,
		TableName: resolveTableName(typ),
	}
	for i := 0; i < typ.NumField(); i++ {
		sf := typ.Field(i)
		if !sf.IsExported() {
			continue
		}
		tag := sf.Tag.Get("orm")
		if tag == "-" {
			continue
		}
		fm := &FieldMeta{
			GoName:     sf.Name,
			Type:       sf.Type,
			FieldIndex: sf.Index,
			Column:     toSnake(sf.Name),
		}
		if tag != "" {
			opts := parseTagOpts(tag)
			if v := opts["column"]; v != "" {
				fm.Column = v
			}
			if _, ok := opts["primaryKey"]; ok {
				fm.IsPrimary = true
			}
			if _, ok := opts["autoIncrement"]; ok {
				fm.IsAutoIncr = true
			}
			if _, ok := opts["nullable"]; ok {
				fm.Nullable = true
			}
			if _, ok := opts["omitempty"]; ok {
				fm.OmitEmpty = true
			}
		}
		if fm.IsPrimary {
			// 主键默认 NOT NULL
			fm.Nullable = false
		}
		meta.Fields = append(meta.Fields, fm)
		meta.Columns = append(meta.Columns, fm.Column)
		if fm.IsPrimary && meta.Primary == nil {
			meta.Primary = fm
		}
	}
	if actual, loaded := modelCache.LoadOrStore(typ, meta); loaded {
		return actual.(*ModelMeta), nil
	}
	return meta, nil
}

func resolveTableName(typ reflect.Type) string {
	tabler := reflect.TypeOf((*Tabler)(nil)).Elem()
	if reflect.PointerTo(typ).Implements(tabler) {
		v := reflect.New(typ)
		out := v.MethodByName("TableName").Call(nil)
		if len(out) == 1 {
			if s, ok := out[0].Interface().(string); ok && s != "" {
				return s
			}
		}
	}
	return toSnake(typ.Name())
}

// toSnake 将驼峰命名转为下划线命名，例如 "UserID" -> "user_id"
func toSnake(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) ||
				(i+1 < len(runes) && unicode.IsLower(runes[i+1]))) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// parseTagOpts 解析 orm tag，格式: "column:name;primaryKey;omitempty"
func parseTagOpts(tag string) map[string]string {
	opts := make(map[string]string)
	for _, part := range strings.Split(tag, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if idx := strings.Index(part, ":"); idx >= 0 {
			opts[strings.TrimSpace(part[:idx])] = strings.TrimSpace(part[idx+1:])
		} else {
			opts[part] = ""
		}
	}
	return opts
}

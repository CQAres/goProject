package orm

import "errors"

// 包内通用错误定义
var (
	ErrNotFound        = errors.New("orm: record not found")
	ErrNotStruct       = errors.New("orm: target must be a struct or struct pointer")
	ErrNotSlice        = errors.New("orm: target must be a slice pointer")
	ErrMissingTable    = errors.New("orm: table name is required")
	ErrUnsupportedDest = errors.New("orm: unsupported destination type")
	ErrMissingWhere    = errors.New("orm: update/delete without where is not allowed")
)

package codec

import "io"

// Header RPC 消息头，包含服务方法名、序列号和错误信息
type Header struct {
	ServiceMethod string // ServiceMethod 调用的服务方法名，格式为 "Service.Method"
	Seq           uint64 // Seq 请求的序列号，用于匹配请求与响应
	Error         string // Error 服务端返回的错误信息，空表示无错误
}

// Codec 编解码器接口，抽象 RPC 消息的读写能力
type Codec interface {
	io.Closer
	ReadHeader(*Header) error         // ReadHeader 读取消息头
	ReadBody(interface{}) error       // ReadBody 读取消息体
	Write(*Header, interface{}) error // Write 写入消息头和消息体
}

// NewCodecFunc 编解码器工厂函数类型，根据连接创建对应的 Codec 实例
type NewCodecFunc func(io.ReadWriteCloser) Codec

// Type 编解码器类型标识
type Type string

const (
	GobType  Type = "application/gob"  // GobType 基于 gob 的编解码器
	JsonType Type = "application/json" // JsonType 基于 json 的编解码器
)

// NewCodecFuncMap 编解码器工厂函数注册表，按 Type 索引
var NewCodecFuncMap map[Type]NewCodecFunc

// init 初始化编解码器注册表，注册已实现的编解码器
func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
	// JsonType codec 尚未实现，待后续添加
	// NewCodecFuncMap[JsonType] = NewJsonCodec
}

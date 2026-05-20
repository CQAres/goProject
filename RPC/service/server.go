package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"goProject/RPC/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

// MagicNumber RPC 协议魔数，用于标识合法的 RPC 连接
const MagicNumber = 0x3bef5c

// Option RPC 连接协商选项，客户端在建立连接时发送
type Option struct {
	MagicNumber int        // MagicNumber 魔数，用于验证连接合法性
	CodecType   codec.Type // CodecType 编解码器类型，决定后续数据的序列化方式
}

// DefaultOption 默认协商选项，使用 Gob 编解码器
var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

// Server RPC 服务端实例
type Server struct{}

// NewServer 创建一个新的 RPC 服务端实例
func NewServer() *Server {
	return &Server{}
}

// DefaultServer 全局默认服务端实例
var DefaultServer = NewServer()

// Accept 监听并接受客户端连接，为每个连接启动独立的处理协程
func (server *Server) Accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept err: %v\n", err)
			continue
		}
		go server.ServeConn(conn)
	}
}

// Accept 使用默认服务端监听并接受客户端连接
func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

// ServeConn 处理单个连接：读取协商选项，验证合法性，创建编解码器并处理请求
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Printf("decode err: %v\n", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("invalid magic number: %x\n", opt.MagicNumber)
		return
	}
	codecFunc := codec.NewCodecFuncMap[opt.CodecType]
	if codecFunc == nil {
		log.Printf("invalid codec type: %s\n", opt.CodecType)
		return
	}
	server.serveCodec(codecFunc(conn))
}

// invalidRequest 无效请求的占位响应体
var invalidRequest = struct{}{}

// serveCodec 循环读取请求并并发处理，所有请求处理完毕后关闭编解码器
func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.header.Error = err.Error()
			server.sendResponse(cc, req.header, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

// request RPC 请求内部结构，封装消息头、参数和返回值
type request struct {
	header       *codec.Header // header 请求消息头
	argv, replyv reflect.Value // argv 请求参数，replyv 响应返回值
}

// Args 加法服务的请求参数
type Args struct {
	Req string
	A   int // A
	B   int // B
}

// readRequestHeader 从编解码器中读取请求消息头
func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

// readRequest 读取完整的 RPC 请求，包含消息头和消息体
func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{header: h}
	// reflect.New(reflect.TypeOf(Args{})) 相当于 new(Args)，创建一个 *Args
	// gob 解码时会将消息体反序列化到这个 *Args 中，填充 A 和 B 字段
	req.argv = reflect.New(reflect.TypeOf(Args{}))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
		return req, err
	}
	// 预先创建 reply 的占位值，类型为 *string（与 handleRequest 中 SetString 对应）
	req.replyv = reflect.New(reflect.TypeOf(""))
	return req, nil
}

// sendResponse 发送 RPC 响应，通过 sending 互斥锁保证并发安全
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

// handleRequest 处理单个 RPC 请求，从参数中取出 A、B 计算求和并返回
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	// req.argv 是 reflect.Value，指向 *Args
	// .Elem() 解引用得到 Args
	// .FieldByName("A") 按字段名取值，.Interface() 转为 interface{} 再断言为 int
	args := req.argv.Elem() // Args 值
	reqSeq := args.FieldByName("Req").Interface().(string)
	a := args.FieldByName("A").Interface().(int)
	b := args.FieldByName("B").Interface().(int)
	log.Printf("handleRequest: reqSeq: %s, A=%d, B=%d", reqSeq, a, b)

	// 计算结果，设置到 replyv
	res := a + b
	result := fmt.Sprintf("reqSeq: %s, result: %d", reqSeq, res)
	req.replyv.Elem().SetString(result)
	server.sendResponse(cc, req.header, req.replyv.Interface(), sending)
}

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
	"strings"
	"sync"
	"time"
)

// MagicNumber RPC 协议魔数，用于标识合法的 RPC 连接
const MagicNumber = 0x3bef5c

// Option RPC 连接协商选项，客户端在建立连接时发送
type Option struct {
	MagicNumber    int        // MagicNumber 魔数，用于验证连接合法性
	CodecType      codec.Type // CodecType 编解码器类型，决定后续数据的序列化方式
	ConnectTimeOut time.Duration
	HandleTimeOut  time.Duration
}

// DefaultOption 默认协商选项，使用 Gob 编解码器
var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeOut: 10 * time.Second,
	HandleTimeOut:  30 * time.Second,
}

// Server RPC 服务端实例
type Server struct {
	serviceMap sync.Map
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

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
		go server.handleRequest(cc, req, sending, wg, time.Second*10)
	}
	wg.Wait()
	_ = cc.Close()
}

// request RPC 请求内部结构，封装消息头、参数和返回值
type request struct {
	header       *codec.Header // header 请求消息头
	argv, replyv reflect.Value // argv 请求参数，replyv 响应返回值
	mtype        *methodType
	svc          *service
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
	req.svc, req.mtype, err = server.findService(req.header.ServiceMethod)
	if err != nil {
		_ = cc.ReadBody(nil) // 必须消费 body，否则 gob 流错位
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReply()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err:", err)
		return req, err
	}

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

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.header.Error = err.Error()
			server.sendResponse(cc, req.header, req.replyv.Interface(), sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.header, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}

	select {
	case <-time.After(timeout):
		req.header.Error = fmt.Sprintf("rpc server: request handle timeout: %s", timeout)
		server.sendResponse(cc, req.header, invalidRequest, sending)

	case <-called:
		<-sent
	}
}

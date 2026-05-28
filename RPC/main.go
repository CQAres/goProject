package main

import (
	"context"
	"goProject/RPC/registry"
	"goProject/RPC/service"
	"goProject/RPC/xclient"
	"net"
	"net/http"
	"sync"
	"time"
)

type Compute int
type Args struct {
	Num1, Num2 int
}

func (c *Compute) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

// 启动 RPC 注册中心
func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

// startServer 启动 RPC 服务端，随机监听可用端口，通过 addr 通道返回监听地址
func startServer(registryAddr string, wg *sync.WaitGroup) {
	var compute Compute
	l, _ := net.Listen("tcp", ":0")
	server := service.NewServer()
	_ = server.Register(&compute)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)
}

func call(registry string) {
	d := xclient.NewRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			xclient.RpcLog(xc, context.Background(), "call", "Compute.Sum", &xclient.Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadcast(registry string) {
	d := xclient.NewRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RoundRobinSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			xclient.RpcLog(xc, context.Background(), "broadcast", "Compute.Sum", &xclient.Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

// main 客户端入口：连接 RPC 服务端，发送协商选项，循环发送请求并接收响应
func main() {
	registryAddr := "http://localhost:9999/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)
}

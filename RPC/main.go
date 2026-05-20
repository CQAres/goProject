package main

import "C"
import (
	"fmt"
	client2 "goProject/RPC/client"
	"goProject/RPC/service"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

// startServer 启动 RPC 服务端，随机监听可用端口，通过 addr 通道返回监听地址
func startServer(addr chan string) {
	// pick a free port
	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", listen.Addr())
	addr <- listen.Addr().String()
	service.Accept(listen)
}

// main 客户端入口：连接 RPC 服务端，发送协商选项，循环发送请求并接收响应
func main() {
	addr := make(chan string)
	go startServer(addr) // 启动服务

	// connection
	client, _ := client2.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)

	// send options
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			request := fmt.Sprintf("go rpc req %d", i)
			log.Println("request:", request)
			args := &service.Args{
				Req: request,
				A:   rand.Intn(100),
				B:   rand.Intn(100),
			}
			var reply string
			if err := client.Call("Func.Sum", args, &reply); err != nil {
				log.Fatal("call Func.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}

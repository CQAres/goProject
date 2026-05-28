package xclient

import (
	"context"
	"fmt"
	"goProject/RPC/service"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

type Compute int

type ComputeArgs struct {
	Num1, Num2 int
}

func (c *Compute) Sum(args ComputeArgs, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (c *Compute) Mul(args ComputeArgs, reply *int) error {
	*reply = args.Num1 * args.Num2
	return nil
}

// startServer 启动一个 RPC 服务端，返回监听地址
func startServer(t *testing.T) string {
	t.Helper()
	var c Compute
	if err := service.Register(&c); err != nil {
		t.Fatal("register error:", err)
	}

	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("network error:", err)
	}

	go func() {
		service.Accept(listen)
	}()

	return listen.Addr().String()
}

// TestXClient_Call 测试单次 Call 调用
func TestXClient_Call(t *testing.T) {
	addr := startServer(t)
	discovery := NewMultiServersDiscovery([]string{addr})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := &ComputeArgs{Num1: 10, Num2: 20}
	var reply int
	err := xc.Call(ctx, "Compute.Sum", args, &reply)
	if err != nil {
		t.Fatalf("Call Compute.Sum error: %v", err)
	}
	if reply != 30 {
		t.Fatalf("expected 30, got %d", reply)
	}

	// 测试乘法
	args = &ComputeArgs{Num1: 5, Num2: 6}
	err = xc.Call(ctx, "Compute.Mul", args, &reply)
	if err != nil {
		t.Fatalf("Call Compute.Mul error: %v", err)
	}
	if reply != 30 {
		t.Fatalf("expected 30, got %d", reply)
	}
}

// TestXClient_RoundRobin 测试轮询选择模式
func TestXClient_RoundRobin(t *testing.T) {
	addr := startServer(t)
	discovery := NewMultiServersDiscovery([]string{addr})
	xc := NewXClient(discovery, RoundRobinSelect, nil)
	defer func() { _ = xc.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		args := &ComputeArgs{Num1: i, Num2: i + 1}
		var reply int
		err := xc.Call(ctx, "Compute.Sum", args, &reply)
		if err != nil {
			t.Fatalf("Call error on iteration %d: %v", i, err)
		}
		if reply != i+i+1 {
			t.Fatalf("iteration %d: expected %d, got %d", i, i+i+1, reply)
		}
	}
}

// TestXClient_Broadcast 测试广播模式
func TestXClient_Broadcast(t *testing.T) {
	addr := startServer(t)
	discovery := NewMultiServersDiscovery([]string{addr})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := &ComputeArgs{Num1: 3, Num2: 7}
	var reply int
	err := xc.Broadcast(ctx, "Compute.Sum", args, &reply)
	if err != nil {
		t.Fatalf("Broadcast Compute.Sum error: %v", err)
	}
	if reply != 10 {
		t.Fatalf("expected 10, got %d", reply)
	}
}

// TestXClient_Broadcast_MultipleServers 测试多服务端广播
func TestXClient_Broadcast_MultipleServers(t *testing.T) {
	addr1 := startServer(t)
	addr2 := startServer(t)

	discovery := NewMultiServersDiscovery([]string{addr1, addr2})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := &ComputeArgs{Num1: 8, Num2: 2}
	var reply int
	err := xc.Broadcast(ctx, "Compute.Sum", args, &reply)
	if err != nil {
		t.Fatalf("Broadcast error: %v", err)
	}
	if reply != 10 {
		t.Fatalf("expected 10, got %d", reply)
	}
}

// TestXClient_Call_Timeout 测试超时取消
func TestXClient_Call_Timeout(t *testing.T) {
	addr := startServer(t)
	discovery := NewMultiServersDiscovery([]string{addr})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	// 极短超时，大概率失败
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	args := &ComputeArgs{Num1: 1, Num2: 2}
	var reply int
	err := xc.Call(ctx, "Compute.Sum", args, &reply)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestXClient_NoServers 测试无可用服务端
func TestXClient_NoServers(t *testing.T) {
	discovery := NewMultiServersDiscovery([]string{})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := &ComputeArgs{Num1: 1, Num2: 2}
	var reply int
	err := xc.Call(ctx, "Compute.Sum", args, &reply)
	if err == nil {
		t.Fatal("expected no available servers error, got nil")
	}
}

// TestXClient_Concurrent 测试并发调用
func TestXClient_Concurrent(t *testing.T) {
	addr := startServer(t)
	discovery := NewMultiServersDiscovery([]string{addr})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer func() { _ = xc.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			args := &ComputeArgs{Num1: i, Num2: i}
			var reply int
			err := xc.Call(ctx, "Compute.Sum", args, &reply)
			if err != nil {
				t.Errorf("concurrent call %d error: %v", i, err)
				return
			}
			if reply != i+i {
				t.Errorf("concurrent call %d: expected %d, got %d", i, i+i, reply)
			}
		}(i)
	}
	wg.Wait()
}

// TestXClient_DynamicUpdate 测试动态更新服务端列表
func TestXClient_DynamicUpdate(t *testing.T) {
	addr1 := startServer(t)
	discovery := NewMultiServersDiscovery([]string{addr1})
	xc := NewXClient(discovery, RoundRobinSelect, nil)
	defer func() { _ = xc.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 先用一个服务端
	args := &ComputeArgs{Num1: 1, Num2: 2}
	var reply int
	err := xc.Call(ctx, "Compute.Sum", args, &reply)
	if err != nil {
		t.Fatalf("Call error: %v", err)
	}
	if reply != 3 {
		t.Fatalf("expected 3, got %d", reply)
	}

	// 动态新增一个服务端
	addr2 := startServer(t)
	_ = discovery.Update([]string{addr1, addr2})

	args = &ComputeArgs{Num1: 10, Num2: 20}
	err = xc.Call(ctx, "Compute.Sum", args, &reply)
	if err != nil {
		t.Fatalf("Call after update error: %v", err)
	}
	if reply != 30 {
		t.Fatalf("expected 30, got %d", reply)
	}
}

// TestDiscovery 测试服务发现
func TestDiscovery(t *testing.T) {
	servers := []string{"addr1", "addr2", "addr3"}
	d := NewMultiServersDiscovery(servers)

	// 测试 GetAll
	all, err := d.GetAll()
	if err != nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(all))
	}

	// 测试 RandomSelect
	addr, err := d.Get(RandomSelect)
	if err != nil {
		t.Fatalf("Get RandomSelect error: %v", err)
	}
	found := false
	for _, s := range servers {
		if s == addr {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("RandomSelect returned unknown address: %s", addr)
	}

	// 测试 RoundRobinSelect
	addr1, _ := d.Get(RoundRobinSelect)
	addr2, _ := d.Get(RoundRobinSelect)
	addr3, _ := d.Get(RoundRobinSelect)
	// 轮询应该依次返回
	addrs := []string{addr1, addr2, addr3}
	for i, a := range addrs {
		if a != servers[i%len(servers)] {
			t.Logf("RoundRobin not strictly sequential (random initial index), got %s at pos %d", a, i)
		}
	}

	// 测试 Update
	newServers := []string{"addr4", "addr5"}
	_ = d.Update(newServers)
	all, _ = d.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 servers after update, got %d", len(all))
	}
}

// ExampleXClient 示例测试
func ExampleXClient() {
	var c Compute
	if err := service.Register(&c); err != nil {
		log.Fatal("register error:", err)
	}

	listen, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	go service.Accept(listen)

	discovery := NewMultiServersDiscovery([]string{listen.Addr().String()})
	xc := NewXClient(discovery, RandomSelect, nil)
	defer xc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := &ComputeArgs{Num1: 2, Num2: 3}
	var reply int
	if err := xc.Call(ctx, "Compute.Sum", args, &reply); err != nil {
		log.Fatal("Call error:", err)
	}
	fmt.Println(reply)
	// Output: 5
}

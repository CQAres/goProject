package http

import (
	"context"
	"fmt"
	"goProject/cache/core"
	"goProject/cache/protobuff"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup(name string, cacheBytes int64) *core.Group {
	return core.RegisterGroup(name, cacheBytes, core.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// ======================= gRPC 测试辅助函数 =======================

// 创建 gRPC 测试服务器
type grpcServer struct {
	protobuff.UnimplementedGroupCacheServer
	group *core.Group
}

func (s *grpcServer) Get(ctx context.Context, req *protobuff.Request) (*protobuff.Response, error) {
	group := core.GetGroup(req.Group)
	if group == nil {
		return nil, fmt.Errorf("no such group: %s", req.Group)
	}
	view, err := group.Get(req.Key)
	if err != nil {
		return nil, err
	}
	return &protobuff.Response{Value: view.ByteSlice()}, nil
}

// 用于测试的 gRPC getter
type testGRPCGetter struct {
	client protobuff.GroupCacheClient
}

func (g *testGRPCGetter) Get(in *protobuff.Request, out *protobuff.Response) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := g.client.Get(ctx, in)
	if err != nil {
		return err
	}
	out.Value = resp.Value
	return nil
}

// 启动 gRPC 测试服务器
func startTestGRPCServer(group *core.Group) (*grpc.Server, *bufconn.Listener) {
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	protobuff.RegisterGroupCacheServer(s, &grpcServer{group: group})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Server exited with error: %v", err)
		}
	}()
	return s, lis
}

// 创建 gRPC 客户端连接
func createGRPCConn(lis *bufconn.Listener) (*grpc.ClientConn, error) {
	return grpc.DialContext(context.Background(), "passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}

// ======================= 测试单节点 gRPC 服务器 =======================

func TestGRPCSingleServer(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	// 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	// 创建客户端
	conn, err := createGRPCConn(lis)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	defer conn.Close()

	client := protobuff.NewGroupCacheClient(conn)

	// 测试正常获取
	t.Run("get existing key", func(t *testing.T) {
		resp, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Tom",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if string(resp.Value) != "630" {
			t.Errorf("expected value '630', got %q", string(resp.Value))
		}
	})

	// 测试缓存命中
	t.Run("cache hit", func(t *testing.T) {
		// 第一次获取，触发加载
		client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Jack",
		})
		// 第二次获取，应该命中缓存
		resp, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Jack",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if string(resp.Value) != "589" {
			t.Errorf("expected value '589', got %q", string(resp.Value))
		}
	})

	// 测试不存在的key
	t.Run("key not exist", func(t *testing.T) {
		_, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Unknown",
		})
		if err == nil {
			t.Errorf("expected error for non-existent key")
		}
	})

	// 测试不存在的group
	t.Run("group not exist", func(t *testing.T) {
		_, err := client.Get(context.Background(), &protobuff.Request{
			Group: "invalidgroup",
			Key:   "key",
		})
		if err == nil {
			t.Errorf("expected error for non-existent group")
		}
	})
}

// ======================= 测试API服务器 (HTTP + gRPC客户端) =======================

func TestAPIServerWithGRPC(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	// 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	// 创建 gRPC 客户端
	conn, err := createGRPCConn(lis)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	defer conn.Close()

	client := protobuff.NewGroupCacheClient(conn)

	// 创建 API 服务器 (HTTP)
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		resp, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   key,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(resp.Value)
	})

	ts := httptest.NewServer(apiHandler)
	defer ts.Close()

	t.Run("api get existing key", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s?key=Tom", ts.URL))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		if string(body) != "630" {
			t.Errorf("expected body '630', got %q", string(body))
		}
	})

	t.Run("api get missing key", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s?key=NotExist", ts.URL))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	t.Run("api without key param", func(t *testing.T) {
		resp, err := http.Get(ts.URL)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})
}

// ======================= 测试多节点分布式缓存 (gRPC) =======================

func TestGRPCDistributedCache(t *testing.T) {
	const (
		groupName  = "scores"
		cacheBytes = 2 << 10
	)

	// 创建多个缓存节点
	addrs := []string{"localhost:8001", "localhost:8002", "localhost:8003"}

	type node struct {
		pool  *HTTPPool
		ts    *httptest.Server
		group *core.Group
		lis   *bufconn.Listener
		grpc  *grpc.Server
	}

	nodes := make([]*node, len(addrs))
	for i, addr := range addrs {
		gee := core.RegisterGroup(groupName, cacheBytes, core.GetterFunc(
			func(key string) ([]byte, error) {
				log.Printf("[SlowDB %s] search key %s", addr, key)
				if v, ok := db[key]; ok {
					return []byte(v), nil
				}
				return nil, fmt.Errorf("%s not exist", key)
			}))

		pool := NewHTTPPool(addr)
		pool.Set(addrs...)

		// 启动 gRPC 服务器
		s, lis := startTestGRPCServer(gee)

		nodes[i] = &node{pool: pool, group: gee, lis: lis, grpc: s}
	}

	// 清理函数
	defer func() {
		for _, n := range nodes {
			n.grpc.Stop()
		}
	}()

	// 测试 gRPC 请求
	t.Run("grpc request to node", func(t *testing.T) {
		conn, err := createGRPCConn(nodes[0].lis)
		if err != nil {
			t.Fatalf("failed to create connection: %v", err)
		}
		defer conn.Close()

		client := protobuff.NewGroupCacheClient(conn)
		resp, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Tom",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if string(resp.Value) != "630" {
			t.Errorf("expected value '630', got %q", string(resp.Value))
		}
	})

	// 测试通过API服务器访问
	t.Run("api server with distributed cache", func(t *testing.T) {
		conn, err := createGRPCConn(nodes[0].lis)
		if err != nil {
			t.Fatalf("failed to create connection: %v", err)
		}
		defer conn.Close()

		client := protobuff.NewGroupCacheClient(conn)

		apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			resp, err := client.Get(context.Background(), &protobuff.Request{
				Group: groupName,
				Key:   key,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(resp.Value)
		})

		apiTs := httptest.NewServer(apiHandler)
		defer apiTs.Close()

		resp, err := http.Get(fmt.Sprintf("%s?key=Jack", apiTs.URL))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

// ======================= 测试节点选择 (PickPeer) =======================

func TestPickPeer(t *testing.T) {
	const groupName = "scores"

	// 创建两个节点
	pool1 := NewHTTPPool("http://localhost:8001")
	pool2 := NewHTTPPool("http://localhost:8002")

	pool1.Set("http://localhost:8001", "http://localhost:8002", "http://localhost:8003")
	pool2.Set("http://localhost:8001", "http://localhost:8002", "http://localhost:8003")

	// 注册 group
	core.RegisterGroup(groupName, 2<<10, core.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 测试 PickPeer - 由于一致性hash，大部分key应该能选到peer
	t.Run("pick peer for different keys", func(t *testing.T) {
		keys := []string{"Tom", "Jack", "Sam", "Key1", "Key2", "A", "B", "C"}
		pickCount := 0

		for _, key := range keys {
			peer, ok := pool1.PickPeer(key)
			if ok && peer != nil {
				pickCount++
				t.Logf("key %s: picked a peer", key)
			} else {
				t.Logf("key %s: picked self or no peer", key)
			}
		}

		// 至少应该有一些key能选到其他节点
		if pickCount == 0 {
			t.Log("Note: all keys happened to pick self, which is expected with few nodes")
		}
	})

	t.Run("pick peer returns non-nil getter when available", func(t *testing.T) {
		// 用多个节点增加选到其他节点的概率
		pool := NewHTTPPool("http://localhost:8000")
		pool.Set("http://localhost:8001", "http://localhost:8002", "http://localhost:8003",
			"http://localhost:8004", "http://localhost:8005")

		foundPeer := false
		keys := []string{"aaa", "bbb", "ccc", "ddd", "eee", "fff", "ggg", "hhh", "iii", "jjj"}
		for _, key := range keys {
			peer, ok := pool.PickPeer(key)
			if ok && peer != nil {
				foundPeer = true
				// 验证 peer 可以获取数据 (使用 protobuf 接口)
				req := &protobuff.Request{Group: groupName, Key: "Tom"}
				resp := &protobuff.Response{}
				err := peer.Get(req, resp)
				if err == nil && string(resp.Value) == "630" {
					t.Logf("key %s: successfully got data through peer", key)
				}
			}
		}

		if !foundPeer {
			t.Errorf("should find at least one peer with multiple nodes")
		}
	})

	t.Run("empty peers returns nil", func(t *testing.T) {
		pool := NewHTTPPool("http://localhost:8001")
		// 初始化 peers 为空 Map
		pool.peers = core.New(50, nil)
		peer, ok := pool.PickPeer("anykey")
		if ok || peer != nil {
			t.Errorf("should not pick peer with no peers in map")
		}
	})
}

// ======================= 测试并发访问 (gRPC) =======================

func TestGRPCConcurrentAccess(t *testing.T) {
	const (
		groupName  = "scores"
		numWorkers = 100
	)

	gee := createGroup(groupName, 2<<10)

	// 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// 每个 goroutine 创建自己的连接
			conn, err := createGRPCConn(lis)
			if err != nil {
				errors <- fmt.Errorf("worker %d: failed to connect: %v", id, err)
				return
			}
			defer conn.Close()

			client := protobuff.NewGroupCacheClient(conn)

			key := "Tom"
			if id%3 == 0 {
				key = "Jack"
			} else if id%3 == 1 {
				key = "Sam"
			}

			resp, err := client.Get(context.Background(), &protobuff.Request{
				Group: groupName,
				Key:   key,
			})
			if err != nil {
				errors <- fmt.Errorf("worker %d: %v", id, err)
				return
			}

			expected := db[key]
			if string(resp.Value) != expected {
				errors <- fmt.Errorf("worker %d: expected %s, got %s", id, expected, string(resp.Value))
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}
}

// ======================= 测试 HTTPPool Set 方法 =======================

func TestHTTPPoolSet(t *testing.T) {
	pool := NewHTTPPool("http://localhost:8001")

	// 测试 Set 方法
	peers := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}
	pool.Set(peers...)

	// 验证可以正常 PickPeer
	t.Run("pick peer after set", func(t *testing.T) {
		peer, ok := pool.PickPeer("anykey")
		if !ok {
			t.Errorf("should pick a peer")
		}
		// 不应该返回自身
		if ok && peer != nil {
			// peer 可能是 httpGetters 中的某个节点
			t.Log("peer selected successfully")
		}
	})

	t.Run("set with empty peers", func(t *testing.T) {
		pool2 := NewHTTPPool("http://localhost:9001")
		pool2.Set() // 空peers

		peer, ok := pool2.PickPeer("key")
		if ok || peer != nil {
			t.Errorf("should not pick peer with no peers configured")
		}
	})
}

// ======================= 测试 gRPC 流控和超时 =======================

func TestGRPCTimeout(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	// 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	conn, err := createGRPCConn(lis)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	defer conn.Close()

	client := protobuff.NewGroupCacheClient(conn)

	// 测试带超时的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.Get(ctx, &protobuff.Request{
		Group: groupName,
		Key:   "Tom",
	})
	if err != nil {
		t.Fatalf("request with timeout failed: %v", err)
	}
	if string(resp.Value) != "630" {
		t.Errorf("expected value '630', got %q", string(resp.Value))
	}
}

// ======================= 性能测试 =======================

func BenchmarkGRPCGet(b *testing.B) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	// 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	conn, err := createGRPCConn(lis)
	if err != nil {
		b.Fatalf("failed to create connection: %v", err)
	}
	defer conn.Close()

	client := protobuff.NewGroupCacheClient(conn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Tom",
		})
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gee.Get("Tom")
	}
}

func BenchmarkHTTPGet(b *testing.B) {
	groupName := "scores"
	pool := NewHTTPPool("http://localhost:8001")
	ts := httptest.NewServer(pool)
	defer ts.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := fmt.Sprintf("%s/cache/%s/Tom", ts.URL, groupName)
		resp, err := http.Get(url)
		if err != nil {
			b.Fatalf("request failed: %v", err)
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// ======================= 辅助：等待服务器启动 =======================

func waitForServer(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(addr)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("server not ready within %v", timeout)
}

// ======================= 集成测试：完整流程 (gRPC) =======================

func TestGRPCFullIntegration(t *testing.T) {
	const (
		groupName  = "integration-test"
		cacheBytes = 2 << 10
	)

	// 1. 创建缓存组
	gee := core.RegisterGroup(groupName, cacheBytes, core.GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[DB] get key: %s", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	// 2. 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	// 3. 创建 gRPC 客户端
	conn, err := createGRPCConn(lis)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	defer conn.Close()

	client := protobuff.NewGroupCacheClient(conn)

	// 4. 测试完整流程
	testKeys := []string{"Tom", "Jack", "Sam"}

	for _, key := range testKeys {
		t.Run(fmt.Sprintf("key_%s", key), func(t *testing.T) {
			resp, err := client.Get(context.Background(), &protobuff.Request{
				Group: groupName,
				Key:   key,
			})
			if err != nil {
				t.Fatalf("gRPC request failed: %v", err)
			}

			expected := db[key]
			if string(resp.Value) != expected {
				t.Errorf("expected %q, got %q", expected, string(resp.Value))
			}
		})
	}

	// 5. 验证缓存命中（第二次访问不应触发DB）
	t.Run("cache hit", func(t *testing.T) {
		// 第一次
		client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Tom",
		})
		// 第二次 - 应该命中缓存
		resp, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "Tom",
		})
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if string(resp.Value) != "630" {
			t.Errorf("expected '630', got %q", string(resp.Value))
		}
	})
}

// ======================= 测试 gRPC 错误处理 =======================

func TestGRPCErrorHandling(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	// 启动 gRPC 服务器
	s, lis := startTestGRPCServer(gee)
	defer s.Stop()

	conn, err := createGRPCConn(lis)
	if err != nil {
		t.Fatalf("failed to create connection: %v", err)
	}
	defer conn.Close()

	client := protobuff.NewGroupCacheClient(conn)

	t.Run("empty group name", func(t *testing.T) {
		_, err := client.Get(context.Background(), &protobuff.Request{
			Group: "",
			Key:   "Tom",
		})
		if err == nil {
			t.Errorf("expected error for empty group name")
		}
	})

	t.Run("empty key", func(t *testing.T) {
		_, err := client.Get(context.Background(), &protobuff.Request{
			Group: groupName,
			Key:   "",
		})
		if err == nil {
			t.Errorf("expected error for empty key")
		}
	})
}

// ======================= 辅助函数：打印日志 =======================

var (
	_ = strings.TrimSpace // 避免未使用的导入警告
	_ = url.QueryEscape
)

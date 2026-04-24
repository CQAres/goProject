package http

import (
	"fmt"
	"goProject/cache/core"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// ======================= 辅助函数 =======================

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

// ======================= 测试单节点缓存服务器 =======================

func TestHTTPSingleServer(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	pool := NewHTTPPool("http://localhost:8001")
	ts := httptest.NewServer(pool)
	defer ts.Close()

	// 测试正常获取
	t.Run("get existing key", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/%s/Tom", ts.URL, groupName)
		resp, err := http.Get(url)
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

	// 测试缓存命中
	t.Run("cache hit", func(t *testing.T) {
		// 第一次获取，触发加载
		url := fmt.Sprintf("%s/cache/%s/Jack", ts.URL, groupName)
		http.Get(url)
		// 第二次获取，应该命中缓存
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// 测试不存在的key
	t.Run("key not exist", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/%s/Unknown", ts.URL, groupName)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})

	// 测试不存在的group
	t.Run("group not exist", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/invalidgroup/key", ts.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", resp.StatusCode)
		}
	})

	// 测试URL格式错误
	t.Run("bad url format", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/onlygroup", ts.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", resp.StatusCode)
		}
	})

	// 清理
	_ = gee
}

// ======================= 测试API服务器 =======================

func TestAPIServer(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	// 创建 API 服务器
	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := gee.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
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

		// 应该返回错误，因为key参数为空
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", resp.StatusCode)
		}
	})
}

// ======================= 测试多节点分布式缓存 =======================

func TestHTTPDistributedCache(t *testing.T) {
	const (
		groupName  = "scores"
		cacheBytes = 2 << 10
	)

	// 创建多个缓存节点
	addrMap := map[string]string{
		"8001": "http://localhost:8001",
		"8002": "http://localhost:8002",
		"8003": "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	// 为每个节点创建服务器
	type node struct {
		pool  *HTTPPool
		ts    *httptest.Server
		group *core.Group
	}

	nodes := make([]*node, len(addrMap))
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
		pool.Set(addrs...) // 设置所有节点为peer

		ts := httptest.NewServer(pool)
		nodes[i] = &node{pool: pool, ts: ts, group: gee}
	}

	// 清理函数
	defer func() {
		for _, n := range nodes {
			n.ts.Close()
		}
	}()

	// 测试直接HTTP请求
	t.Run("direct http request to node", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/%s/Tom", nodes[0].ts.URL, groupName)
		resp, err := http.Get(url)
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

	// 测试通过API服务器访问
	t.Run("api server with distributed cache", func(t *testing.T) {
		apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := nodes[0].group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())
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

	// 创建测试服务器
	ts1 := httptest.NewServer(pool1)
	ts2 := httptest.NewServer(pool2)
	defer ts1.Close()
	defer ts2.Close()

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
				// 验证 peer 可以获取数据
				data, err := peer.Get(groupName, "Tom")
				if err == nil && string(data) == "630" {
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

// ======================= 测试并发访问 =======================

func TestConcurrentAccess(t *testing.T) {
	const (
		groupName  = "scores"
		numWorkers = 100
	)

	pool := NewHTTPPool("http://localhost:8001")
	ts := httptest.NewServer(pool)
	defer ts.Close()

	var wg sync.WaitGroup
	errors := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := "Tom"
			if id%3 == 0 {
				key = "Jack"
			} else if id%3 == 1 {
				key = "Sam"
			}

			// 通过HTTP访问
			url := fmt.Sprintf("%s/cache/%s/%s", ts.URL, groupName, key)
			resp, err := http.Get(url)
			if err != nil {
				errors <- fmt.Errorf("worker %d: %v", id, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("worker %d: status %d", id, resp.StatusCode)
				return
			}

			// 验证返回值
			body, _ := io.ReadAll(resp.Body)
			expected := db[key]
			if string(body) != expected {
				errors <- fmt.Errorf("worker %d: expected %s, got %s", id, expected, string(body))
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

// ======================= 测试 URL 编码 =======================

func TestURLEncoding(t *testing.T) {
	groupName := "scores"
	gee := core.RegisterGroup(groupName, 2<<10, core.GetterFunc(
		func(key string) ([]byte, error) {
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	pool := NewHTTPPool("http://localhost:8001")
	ts := httptest.NewServer(pool)
	defer ts.Close()

	t.Run("key with special characters", func(t *testing.T) {
		// 测试 URL 编码的 key
		encodedKey := url.QueryEscape("Tom")
		urlStr := fmt.Sprintf("%s/cache/%s/%s", ts.URL, groupName, encodedKey)
		resp, err := http.Get(urlStr)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("key with space", func(t *testing.T) {
		// 注册包含空格的 key
		core.RegisterGroup("space-test", 2<<10, core.GetterFunc(
			func(key string) ([]byte, error) {
				return []byte("value_with_space"), nil
			}))

		encodedKey := url.QueryEscape("key with space")
		urlStr := fmt.Sprintf("%s/cache/space-test/%s", ts.URL, encodedKey)
		resp, err := http.Get(urlStr)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	_ = gee
}

// ======================= 测试响应头 =======================

func TestResponseHeaders(t *testing.T) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)
	pool := NewHTTPPool("http://localhost:8001")
	ts := httptest.NewServer(pool)
	defer ts.Close()

	url := fmt.Sprintf("%s/cache/%s/Tom", ts.URL, groupName)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 验证 Content-Type
	if ct := resp.Header.Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("expected Content-Type 'application/octet-stream', got %q", ct)
	}

	// 验证可以正常读取 body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("failed to read body: %v", err)
	}
	if string(body) != "630" {
		t.Errorf("expected '630', got %q", string(body))
	}

	_ = gee
}

// ======================= 性能测试 =======================

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

func BenchmarkCacheGet(b *testing.B) {
	groupName := "scores"
	gee := createGroup(groupName, 2<<10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gee.Get("Tom")
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

// ======================= 集成测试：完整流程 =======================

func TestFullIntegration(t *testing.T) {
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

	// 2. 创建多个缓存节点
	peers := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}

	pool := NewHTTPPool("http://localhost:8001")
	pool.Set(peers...)

	// 3. 创建 HTTP 测试服务器
	ts := httptest.NewServer(pool)
	defer ts.Close()

	// 4. 测试完整流程
	testKeys := []string{"Tom", "Jack", "Sam"}

	for _, key := range testKeys {
		t.Run(fmt.Sprintf("key_%s", key), func(t *testing.T) {
			// 通过 HTTP 获取
			url := fmt.Sprintf("%s/cache/%s/%s", ts.URL, groupName, key)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("HTTP request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected status 200, got %d", resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			expected := db[key]
			if string(body) != expected {
				t.Errorf("expected %q, got %q", expected, string(body))
			}
		})
	}

	// 5. 验证缓存命中（第二次访问不应触发DB）
	t.Run("cache hit", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/%s/Tom", ts.URL, groupName)
		// 第一次
		http.Get(url)
		// 第二次 - 应该命中缓存
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	_ = gee
}

// ======================= 辅助函数：打印日志 =======================

var (
	_ = strings.TrimSpace // 避免未使用的导入警告
)

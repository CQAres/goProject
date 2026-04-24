package core

import (
	"fmt"
	"sync"
	"testing"
)

func TestGroupGet(t *testing.T) {
	testData := map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	loadCounts := make(map[string]int, len(testData))
	group := RegisterGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			// 模拟 db 耗时操作
			fmt.Println("[SlowDB] search key", key)
			if v, ok := testData[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	for k, v := range testData {
		if view, err := group.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of Tom")
		}
		if _, err := group.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}

// 测试空key
func TestGroupEmptyKey(t *testing.T) {
	group := RegisterGroup("test", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			return []byte(key), nil
		}))

	_, err := group.Get("")
	if err == nil {
		t.Errorf("should return error for empty key")
	}
}

// 测试获取不存在的Group
func TestGetGroupNotExist(t *testing.T) {
	g := GetGroup("notexist")
	if g != nil {
		t.Errorf("should return nil for non-existent group")
	}
}

// 测试重复注册同名Group会panic
func TestRegisterGroupDuplicate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic when registering duplicate group")
		}
	}()

	RegisterGroup("duplicate", 2<<10, nil)
	RegisterGroup("duplicate", 2<<10, nil) // 这里会panic
}

// 测试并发访问同一个Group
func TestGroupConcurrentGet(t *testing.T) {
	group := RegisterGroup("concurrent", 2<<20, GetterFunc(
		func(key string) ([]byte, error) {
			return []byte("value"), nil
		}))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			view, err := group.Get("key")
			if err != nil {
				t.Errorf("get error: %v", err)
			}
			if view.String() != "value" {
				t.Errorf("wrong value: %s", view.String())
			}
		}()
	}
	wg.Wait()
}

// 测试Group缓存未命中后加载
func TestGroupCacheMissThenLoad(t *testing.T) {
	loadCount := 0
	group := RegisterGroup("loadtest", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			loadCount++
			return []byte("loaded_value"), nil
		}))

	// 第一次访问，应该从getter加载
	view1, err := group.Get("testkey")
	if err != nil || view1.String() != "loaded_value" {
		t.Errorf("first get failed")
	}
	if loadCount != 1 {
		t.Errorf("loadCount should be 1, got %d", loadCount)
	}

	// 第二次访问，应该命中缓存
	_, err = group.Get("testkey")
	if loadCount != 1 {
		t.Errorf("loadCount should still be 1 after cache hit, got %d", loadCount)
	}
}

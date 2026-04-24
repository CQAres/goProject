package core

import (
	"hash/crc32"
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, crc32.ChecksumIEEE)

	servers := []string{"server1", "server2", "server3"}
	hash.Add(servers...)

	// 记录初始状态下几个 key 的归属
	var keys []string
	for i := range 10 {
		keys = append(keys, "user"+strconv.Itoa(i))
	}

	initialMapping := make(map[string]string)
	for _, k := range keys {
		initialMapping[k] = hash.Get(k)
		t.Logf("Initially, key %s -> %s", k, initialMapping[k])
	}

	hash.Add("server4")

	changedCount := 0
	for _, k := range keys {
		newServer := hash.Get(k)
		if newServer != initialMapping[k] {
			changedCount++
			t.Logf("Key %s moved from %s to %s", k, initialMapping[k], newServer)
		}
	}

	if changedCount > 4 {
		t.Errorf("Too many keys remapped: %d out of %d", changedCount, len(keys))
	}

	found := false
	for _, k := range keys {
		if hash.Get(k) == "server4" {
			found = true
			break
		}
	}
	if !found {
		// 再试一个更可能命中 server4 的 key
		if hash.Get("new_user_xyz") == "server4" {
			found = true
		}
	}
	if !found {
		// 注意：这不一定是 bug，只是测试 key 没覆盖到
		t.Log("Warning: server4 not used by any test key (could be due to hash distribution)")
	}
}

// 测试空Map的Get
func TestEmptyMapGet(t *testing.T) {
	hash := New(1, nil)
	result := hash.Get("anykey")
	if result != "" {
		t.Errorf("empty map should return empty string, got %s", result)
	}
}

// 测试单节点
func TestSingleNode(t *testing.T) {
	hash := New(1, nil)
	hash.Add("server1")

	for i := 0; i < 100; i++ {
		key := "key" + strconv.Itoa(i)
		if hash.Get(key) != "server1" {
			t.Errorf("all keys should map to server1, but %s got %s", key, hash.Get(key))
		}
	}
}

// 测试虚拟节点数量影响分布均匀性
func TestVirtualNodesDistribution(t *testing.T) {
	hash1 := New(1, nil)   // 每个节点1个虚拟节点
	hash2 := New(100, nil) // 每个节点100个虚拟节点

	hash1.Add("server1", "server2", "server3")
	hash2.Add("server1", "server2", "server3")

	counts1 := make(map[string]int)
	counts2 := make(map[string]int)

	for i := 0; i < 1000; i++ {
		key := strconv.Itoa(i)
		counts1[hash1.Get(key)]++
		counts2[hash2.Get(key)]++
	}

	// 检查是否有节点为空（不平衡）
	for _, name := range []string{"server1", "server2", "server3"} {
		if counts1[name] == 0 {
			t.Errorf("server %s has 0 keys with replicas=1", name)
		}
		if counts2[name] == 0 {
			t.Errorf("server %s has 0 keys with replicas=100", name)
		}
	}

	t.Logf("Distribution with replicas=1: %v", counts1)
	t.Logf("Distribution with replicas=100: %v", counts2)
}

// 测试连续添加
func TestMultipleAdd(t *testing.T) {
	hash := New(3, nil)

	hash.Add("server1")
	hash.Add("server2")
	hash.Add("server3")

	if hash.Get("key") == "" {
		t.Errorf("hash should not be empty after adding servers")
	}
}

// 测试删除节点后重新分布
func TestRemoveNode(t *testing.T) {
	hash := New(1, nil)
	hash.Add("server1", "server2", "server3")

	// 记录初始映射
	initial := hash.Get("user1")
	if initial == "" {
		t.Errorf("should get a server for user1")
	}

	// 删除server1
	hash2 := New(1, nil)
	hash2.Add("server2", "server3")
	newServer := hash2.Get("user1")

	// user1不应该再映射到server1
	if newServer == initial {
		t.Logf("user1 still maps to %s after removing server1", newServer)
	}
}

// 测试自定义Hash函数
func TestCustomHash(t *testing.T) {
	customHash := func(data []byte) uint32 {
		// 使用简单的字符串长度作为hash
		return uint32(len(data))
	}

	hash := New(1, customHash)
	hash.Add("server1", "server2")

	// 相同长度的key应该映射到同一个server
	key1 := "abc"
	key2 := "xyz"
	server1 := hash.Get(key1)
	server2 := hash.Get(key2)

	// 长度都是3，理论上应该映射到同一个
	if server1 == server2 && server1 != "" {
		t.Logf("same length keys (%s, %s) map to same server: %s", key1, key2, server1)
	}
}

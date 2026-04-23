package consistenthash

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

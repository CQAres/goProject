package cache

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTP(t *testing.T) {
	http_db := map[string]string{
		"alice":   "630",
		"bob":     "589",
		"charlie": "567",
	}

	groupName := "test-group"
	RegisterGroup(groupName, 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := http_db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	pool := NewHTTPPool("test-node")

	ts := httptest.NewServer(pool)
	defer ts.Close()

	// 构造测试用例
	tests := []struct {
		name           string
		key            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "existing key",
			key:            "alice",
			expectedStatus: http.StatusOK,
			expectedBody:   "630",
		},
		{
			name:           "another existing key",
			key:            "bob",
			expectedStatus: http.StatusOK,
			expectedBody:   "589",
		},
		{
			name:           "another existing key",
			key:            "charlie",
			expectedStatus: http.StatusOK,
			expectedBody:   "567",
		},
		{
			name:           "nonexistent key",
			key:            "nonexist",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "", // 不检查内容
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("%s/cache/%s/%s", ts.URL, groupName, tt.key)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body := make([]byte, len(tt.expectedBody))
				n, _ := resp.Body.Read(body)
				got := string(body[:n])
				if got != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, got)
				}
			}
		})
	}

	t.Run("nonexistent group", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/invalidgroup/key", ts.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 for invalid group, got %d", resp.StatusCode)
		}
	})

	t.Run("bad url format", func(t *testing.T) {
		url := fmt.Sprintf("%s/cache/onlygroup", ts.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for bad format, got %d", resp.StatusCode)
		}
	})
}

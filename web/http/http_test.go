package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============ Context Tests ============

func TestNewContext(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello", nil)
	c := newContext(w, req)

	if c.Method != "GET" {
		t.Errorf("expected method GET, got %s", c.Method)
	}
	if c.Path != "/hello" {
		t.Errorf("expected path /hello, got %s", c.Path)
	}
	if c.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", c.StatusCode)
	}
	if c.index != -1 {
		t.Errorf("expected index -1, got %d", c.index)
	}
}

func TestContextQuery(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/search?name=Alice&age=30", nil)
	c := newContext(w, req)

	if c.Query("name") != "Alice" {
		t.Errorf("expected name=Alice, got %s", c.Query("name"))
	}
	if c.Query("age") != "30" {
		t.Errorf("expected age=30, got %s", c.Query("age"))
	}
	if c.Query("nonexistent") != "" {
		t.Errorf("expected empty string for nonexistent key")
	}
}

func TestContextPostForm(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/submit", strings.NewReader("username=bob&password=secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = req.ParseForm()
	c := newContext(w, req)

	if c.PostForm("username") != "bob" {
		t.Errorf("expected username=bob, got %s", c.PostForm("username"))
	}
	if c.PostForm("password") != "secret" {
		t.Errorf("expected password=secret, got %s", c.PostForm("password"))
	}
}

func TestContextParam(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/123", nil)
	c := newContext(w, req)
	c.Params = map[string]string{"id": "123"}

	if c.Param("id") != "123" {
		t.Errorf("expected param id=123, got %s", c.Param("id"))
	}
	if c.Param("nonexistent") != "" {
		t.Errorf("expected empty string for nonexistent param")
	}
}

func TestContextStatus(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContext(w, req)

	c.Status(404)
	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
	if c.StatusCode != 404 {
		t.Errorf("expected context status 404, got %d", c.StatusCode)
	}
}

func TestContextSetHeader(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContext(w, req)

	c.SetHeader("X-Custom-Header", "test-value")
	if w.Header().Get("X-Custom-Header") != "test-value" {
		t.Errorf("expected header value test-value, got %s", w.Header().Get("X-Custom-Header"))
	}
}

func TestContextString(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContext(w, req)

	c.String(200, "Hello %s", "World")
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("expected body 'Hello World', got '%s'", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type text/plain, got %s", w.Header().Get("Content-Type"))
	}
}

func TestContextHTML(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContext(w, req)

	c.HTML(200, "<html><body>Hello</body></html>")
	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "<html><body>Hello</body></html>" {
		t.Errorf("expected HTML body, got '%s'", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "text/html" {
		t.Errorf("expected Content-Type text/html, got %s", w.Header().Get("Content-Type"))
	}
}

func TestContextFail(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContext(w, req)

	c.Fail(500, "Internal Server Error")
	if w.Code != 500 {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	if w.Body.String() != "Internal Server Error" {
		t.Errorf("expected error message, got '%s'", w.Body.String())
	}
}

func TestContextNext(t *testing.T) {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	c := newContext(w, req)

	order := make([]int, 0)
	c.handlers = []HandlerFunc{
		func(c *Context) {
			order = append(order, 1)
			c.Next()
			order = append(order, 6)
		},
		func(c *Context) {
			order = append(order, 2)
			c.Next()
			order = append(order, 5)
		},
		func(c *Context) {
			order = append(order, 3)
			c.Next()
			order = append(order, 4)
		},
	}

	c.Next()

	expected := []int{1, 2, 3, 4, 5, 6}
	if len(order) != len(expected) {
		t.Errorf("expected %d calls, got %d", len(expected), len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: expected %d, got %d", i, v, order[i])
		}
	}
}

// ============ Engine Tests ============

func TestEngineBasicRoute(t *testing.T) {
	engine := NewEngine()
	engine.GET("/hello", func(c *Context) {
		c.String(200, "Hello World")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("expected body 'Hello World', got '%s'", w.Body.String())
	}
}

func TestEngineNotFound(t *testing.T) {
	engine := NewEngine()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/notfound", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected status 404, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "404") {
		t.Errorf("expected 404 in response body, got '%s'", w.Body.String())
	}
}

func TestEngineMultipleRoutes(t *testing.T) {
	engine := NewEngine()

	engine.GET("/hello", func(c *Context) {
		c.String(200, "Hello")
	})
	engine.GET("/world", func(c *Context) {
		c.String(200, "World")
	})
	engine.POST("/submit", func(c *Context) {
		c.String(200, "Submitted")
	})

	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{"GET", "/hello", "Hello", 200},
		{"GET", "/world", "World", 200},
		{"POST", "/submit", "Submitted", 200},
		{"DELETE", "/hello", "", 404},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		engine.ServeHTTP(w, req)

		if w.Code != tt.status {
			t.Errorf("%s %s: expected status %d, got %d", tt.method, tt.path, tt.status, w.Code)
		}
		if tt.body != "" && w.Body.String() != tt.body {
			t.Errorf("%s %s: expected body '%s', got '%s'", tt.method, tt.path, tt.body, w.Body.String())
		}
	}
}

// ============ Router Tests ============

func TestParsePattern(t *testing.T) {
	tests := []struct {
		pattern  string
		expected []string
	}{
		{"/", nil},
		{"/user", []string{"user"}},
		{"/user/profile", []string{"user", "profile"}},
		{"/user/:id", []string{"user", ":id"}},
		{"/user/:id/profile", []string{"user", ":id", "profile"}},
		{"/static/*filepath", []string{"static", "*filepath"}},
		{"/a/b/c/d", []string{"a", "b", "c", "d"}},
	}

	for _, tt := range tests {
		result := parsePattern(tt.pattern)
		if tt.expected == nil {
			if result != nil && len(result) > 0 {
				t.Errorf("parsePattern(%s): expected nil, got %v", tt.pattern, result)
			}
		} else {
			if len(result) != len(tt.expected) {
				t.Errorf("parsePattern(%s): expected %v, got %v", tt.pattern, tt.expected, result)
				continue
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("parsePattern(%s): expected %v, got %v", tt.pattern, tt.expected, result)
					break
				}
			}
		}
	}
}

func TestRouterStaticRoute(t *testing.T) {
	engine := NewEngine()
	engine.GET("/user/profile", func(c *Context) {
		c.String(200, "User Profile")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/profile", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "User Profile" {
		t.Errorf("expected 'User Profile', got '%s'", w.Body.String())
	}
}

func TestRouterParamRoute(t *testing.T) {
	engine := NewEngine()
	engine.GET("/user/:id", func(c *Context) {
		id := c.params("id")
		c.String(200, "User ID: %s", id)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/123", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "User ID: 123" {
		t.Errorf("expected 'User ID: 123', got '%s'", w.Body.String())
	}
}

func TestRouterWildcardRoute(t *testing.T) {
	engine := NewEngine()
	engine.GET("/static/*filepath", func(c *Context) {
		path := c.params("filepath")
		c.String(200, "File: %s", path)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/static/css/style.css", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "File: css/style.css" {
		t.Errorf("expected 'File: css/style.css', got '%s'", w.Body.String())
	}
}

func TestRouterMultipleParamRoutes(t *testing.T) {
	engine := NewEngine()
	engine.GET("/user/:uid/post/:pid", func(c *Context) {
		uid := c.params("uid")
		pid := c.params("pid")
		c.String(200, "User %s, Post %s", uid, pid)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/user/456/post/789", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "User 456, Post 789" {
		t.Errorf("expected 'User 456, Post 789', got '%s'", w.Body.String())
	}
}

// ============ RouterGroup Tests ============

func TestRouterGroupBasic(t *testing.T) {
	engine := NewEngine()
	group := engine.Group("/api")

	group.GET("/hello", func(c *Context) {
		c.String(200, "Hello from group")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/hello", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "Hello from group" {
		t.Errorf("expected 'Hello from group', got '%s'", w.Body.String())
	}
}

func TestRouterGroupNested(t *testing.T) {
	engine := NewEngine()
	api := engine.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")

	users.GET("/:id", func(c *Context) {
		id := c.params("id")
		c.String(200, "User ID: %s", id)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "User ID: 123" {
		t.Errorf("expected 'User ID: 123', got '%s'", w.Body.String())
	}
}

func TestRouterGroupMultipleRoutes(t *testing.T) {
	engine := NewEngine()
	group := engine.Group("/api")

	group.GET("/users", func(c *Context) {
		c.String(200, "User List")
	})
	group.GET("/posts", func(c *Context) {
		c.String(200, "Post List")
	})
	group.POST("/posts", func(c *Context) {
		c.String(201, "Post Created")
	})

	tests := []struct {
		method string
		path   string
		body   string
		status int
	}{
		{"GET", "/api/users", "User List", 200},
		{"GET", "/api/posts", "Post List", 200},
		{"POST", "/api/posts", "Post Created", 201},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.path, nil)
		engine.ServeHTTP(w, req)

		if w.Code != tt.status {
			t.Errorf("%s %s: expected status %d, got %d", tt.method, tt.path, tt.status, w.Code)
		}
		if w.Body.String() != tt.body {
			t.Errorf("%s %s: expected body '%s', got '%s'", tt.method, tt.path, tt.body, w.Body.String())
		}
	}
}

// ============ Middleware Tests ============

func TestRecoverMiddleware(t *testing.T) {
	engine := NewEngine()
	engine.Use(Recover())

	engine.GET("/panic", func(c *Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)

	// Should not panic, should return 500
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("expected error message in body, got '%s'", w.Body.String())
	}
}

func TestMiddlewareChain(t *testing.T) {
	engine := NewEngine()

	order := make([]string, 0)

	engine.Use(func(c *Context) {
		order = append(order, "middleware1-start")
		c.Next()
		order = append(order, "middleware1-end")
	})

	engine.Use(func(c *Context) {
		order = append(order, "middleware2-start")
		c.Next()
		order = append(order, "middleware2-end")
	})

	engine.GET("/test", func(c *Context) {
		order = append(order, "handler")
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	expected := []string{
		"middleware1-start",
		"middleware2-start",
		"handler",
		"middleware2-end",
		"middleware1-end",
	}

	if len(order) != len(expected) {
		t.Errorf("expected %d operations, got %d", len(expected), len(order))
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: expected '%s', got '%s'", i, v, order[i])
		}
	}
}

func TestGroupMiddleware(t *testing.T) {
	engine := NewEngine()

	order := make([]string, 0)

	engine.Use(func(c *Context) {
		order = append(order, "global")
		c.Next()
	})

	api := engine.Group("/api")
	api.Use(func(c *Context) {
		order = append(order, "api-group")
		c.Next()
	})

	api.GET("/test", func(c *Context) {
		order = append(order, "handler")
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/test", nil)
	engine.ServeHTTP(w, req)

	expected := []string{"global", "api-group", "handler"}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("at index %d: expected '%s', got '%s'", i, v, order[i])
		}
	}
}

// ============ Response Format Tests ============

func TestResponseContentType(t *testing.T) {
	engine := NewEngine()

	engine.GET("/json", func(c *Context) {
		c.SetHeader("Content-Type", "application/json")
		c.Status(200)
		_, _ = c.Writer.Write([]byte(`{"name":"test"}`))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/json", nil)
	engine.ServeHTTP(w, req)

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestJSONResponse(t *testing.T) {
	engine := NewEngine()

	engine.GET("/api/data", func(c *Context) {
		data := map[string]string{"key": "value"}
		jsonBytes, _ := json.Marshal(data)
		c.SetHeader("Content-Type", "application/json")
		_, _ = c.Writer.Write(jsonBytes)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/data", nil)
	engine.ServeHTTP(w, req)

	var result map[string]string
	json.Unmarshal(w.Body.Bytes(), &result)

	if result["key"] != "value" {
		t.Errorf("expected key=value, got %v", result)
	}
}

// ============ Edge Cases ============

func TestEmptyPath(t *testing.T) {
	engine := NewEngine()
	engine.GET("/", func(c *Context) {
		c.String(200, "Root")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	engine.ServeHTTP(w, req)

	if w.Code != 200 || w.Body.String() != "Root" {
		t.Errorf("expected 200 and 'Root', got %d and '%s'", w.Code, w.Body.String())
	}
}

func TestTrailingSlash(t *testing.T) {
	engine := NewEngine()
	engine.GET("/hello", func(c *Context) {
		c.String(200, "Hello")
	})

	// parsePattern strips empty segments, so /hello/ is treated as /hello
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/hello/", nil)
	engine.ServeHTTP(w, req)

	// This is expected behavior: trailing slash matches the same route
	if w.Code != 200 {
		t.Errorf("expected 200 for /hello/, got %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	engine := NewEngine()
	engine.GET("/resource", func(c *Context) {
		c.String(200, "GET Resource")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/resource", nil)
	engine.ServeHTTP(w, req)

	// Currently returns 404 for method mismatch
	if w.Code != 404 {
		t.Errorf("expected 404 for POST /resource, got %d", w.Code)
	}
}

func TestConcurrentRequests(t *testing.T) {
	engine := NewEngine()
	engine.GET("/test", func(c *Context) {
		c.String(200, "OK")
	})

	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			engine.ServeHTTP(w, req)
			if w.Code == 200 {
				done <- true
			}
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}

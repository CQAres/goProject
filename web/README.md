# GeeWeb HTTP 框架

GeeWeb 是一个轻量级的 Go Web 框架，采用 Gin 风格的 API 设计，支持路由分组、中间件和动态路由参数。

## 项目架构

```
web/http/
├── context.go      # HTTP 上下文封装
├── recover.go      # 错误恢复中间件
├── router.go       # 路由管理
├── routerGroup.go  # 路由分组
├── webHttp.go      # HTTP 引擎（核心）
└── trie/
    └── trie.go     # 前缀树路由匹配
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 1. context.go - HTTP 上下文

### `Context` 结构体

HTTP 请求上下文，封装了请求和响应的相关信息。

```go
type Context struct {
	Writer     http.ResponseWriter  // HTTP 响应写入器
	Req        *http.Request        // HTTP 请求
	Path       string               // 请求路径
	Method     string               // 请求方法
	Params     map[string]string    // 动态路由参数
	StatusCode int                  // 响应状态码
	handlers   []HandlerFunc        // 中间件和处理函数链
	index      int                  // 当前执行的 handler 索引
}
```

| 字段 | 类型 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `Writer` | `http.ResponseWriter` | 用于写入 HTTP 响应 |
| `Req` | `*http.Request` | 原始 HTTP 请求 |
| `Path` | `string` | 请求路径（URL path）|
| `Method` | `string` | HTTP 方法（GET/POST 等）|
| `Params` | `map[string]string` | 动态路由参数，如 `:id` |
| `StatusCode` | `int` | 响应状态码 |
| `handlers` | `[]HandlerFunc` | 中间件和处理函数链 |
| `index` | `int` | 当前执行的 handler 索引，-1 表示开始 |

**特点：**
- **责任链模式**：通过 `handlers` 链和 `index` 实现中间件链式调用
- **参数提取**：支持从 URL 和表单中提取参数
- **响应封装**：提供便捷的响应方法

### `HandlerFunc` 类型

处理函数类型定义。

```go
type HandlerFunc func(*Context)
```

### 核心方法

| 方法 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `Next()` | 执行下一个 handler，实现中间件链 |
| `Query(key string)` | 获取 URL 查询参数 |
| `PostForm(key string)` | 获取 POST 表单参数 |
| `Param(key string)` | 获取动态路由参数（公共方法） |
| `params(key string)` | 获取动态路由参数（私有方法，内部使用） |
| `Status(code int)` | 设置响应状态码 |
| `SetHeader(key, value string)` | 设置响应头 |
| `String(code int, format string, values ...)` | 返回文本响应 |
| `HTML(code int, html string)` | 返回 HTML 响应 |
| `Fail(code int, msg string)` | 返回错误响应 |

### 请求参数获取流程

```
Query("name")     ──▶  req.URL.Query().Get("name")
PostForm("name")  ──▶  req.PostForm.Get("name")
Param("id")       ──▶  Params["id"] (从路由中提取)
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 2. recover.go - 错误恢复中间件

### `Recover` 函数

创建错误恢复中间件，捕获 panic 并返回 500 错误。

```go
func Recover() HandlerFunc
```

**特点：**
- **Panic 恢复**：捕获所有未处理的 panic
- **堆栈跟踪**：记录完整的调用堆栈信息
- **优雅降级**：向客户端返回 500 错误页面
- **安全写入**：恢复后直接操作 ResponseWriter 写入响应，避免级联 panic

**使用示例：**

```go
engine.Use(http.Recover())
```

### `trace` 函数

生成调用堆栈跟踪信息。

```go
func trace(message string) string
```

**输出格式：**
```
错误信息
Traceback:
    /path/to/file1.go:123
    /path/to/file2.go:456
    ...
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 3. trie/trie.go - 前缀树路由匹配

### `Node` 结构体

前缀树节点，用于存储路由模式。

```go
type Node struct {
	Pattern  string    // 完整路由模式（如 "/user/:id/profile"）
	Part     string    // 当前节点的部分（如 "user" 或 ":id"）
	Children []*Node   // 子节点列表
	IsWild   bool      // 是否为通配符节点（: 或 * 开头）
}
```

| 字段 | 类型 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `Pattern` | `string` | 完整路由模式，仅叶子节点有效 |
| `Part` | `string` | 当前段的名称或参数名 |
| `Children` | `[]*Node` | 子节点列表 |
| `IsWild` | `bool` | 是否为通配符（`:id` 或 `*path`）|

### 路由模式支持

| 模式 | 示例 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| 静态 | `/user/profile` | 精确匹配 |
| 参数 | `/user/:id` | 匹配单个路径段 |
| 通配 | `/static/*filepath` | 匹配剩余所有路径 |

### 核心方法

| 方法 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `Insert(pattern, parts, height)` | 插入路由模式 |
| `Search(parts, height)` | 搜索匹配的节点 |
| `MatchChild(part)` | 匹配单个子节点 |
| `MatchChildren(part)` | 匹配所有符合条件的子节点 |

### 路由匹配算法

```
插入 "/user/:id/profile":
root ─▶ "user" ─▶ ":id" ─▶ "profile"

搜索 "/user/123/profile":
1. 匹配 "user"  ─▶ 进入下一层
2. 匹配 ":id"   ─▶ 123 作为 id 参数
3. 匹配 "profile" ─▶ 找到 Pattern
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 4. router.go - 路由管理

### `router` 结构体

路由管理器，使用前缀树存储路由。

```go
type router struct {
	roots    map[string]*trie.Node   // 按方法存储的根节点
	handlers map[string]HandlerFunc  // 路由模式到处理函数的映射
}
```

| 字段 | 类型 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `roots` | `map[string]*trie.Node` | 每种 HTTP 方法对应一棵 trie 树 |
| `handlers` | `map[string]HandlerFunc` | 路由到处理函数的映射 |

### `parsePattern` 函数

解析路由模式为路径段数组。

```go
func parsePattern(pattern string) []string
```

**示例：**

| 输入 | 输出 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `/user/:id/profile` | `["user", ":id", "profile"]` |
| `/static/*filepath` | `["static", "*filepath"]` |

### 核心方法

| 方法 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `addRoute(method, pattern, handler)` | 注册路由 |
| `getRouter(method, path)` | 查找匹配的路由和参数 |
| `handle(c *Context)` | 处理请求，分发到对应 handler |

### 参数提取逻辑

```go
// 路径参数提取
for i, part := range parts {
    if part[0] == ':' {           // 参数：/user/:id
        params[part[1:]] = searchParts[i]
    }
    if part[0] == '*' && len(part) > 1 {  // 通配符：/static/*filepath
        params[part[1:]] = strings.Join(searchParts[i:], "/")
        break
    }
}
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 5. routerGroup.go - 路由分组

### `RouterGroup` 结构体

路由分组，支持前缀和中间件。

```go
type RouterGroup struct {
	prefix      string            // 路由前缀
	middlewares []HandlerFunc     // 属于该分组的中间件
	parent      *RouterGroup      // 父分组
	engine      *Engine           // 所属引擎
}
```

| 字段 | 类型 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `prefix` | `string` | 分组路由前缀，如 `/api/v1` |
| `middlewares` | `[]HandlerFunc` | 该分组使用的中间件 |
| `parent` | `*RouterGroup` | 父分组（支持嵌套分组）|
| `engine` | `*Engine` | 所属的 Engine 实例 |

### 核心方法

| 方法 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `Group(prefix string)` | 创建子分组（不继承父分组中间件） |
| `GET(pattern, handler)` | 注册 GET 请求 |
| `POST(pattern, handler)` | 注册 POST 请求 |
| `addRoute(method, pattern, handler)` | 通用路由注册 |
| `Use(middlewares ...HandlerFunc)` | 添加中间件 |

### 分组嵌套示例

```go
engine := NewEngine()

// 根分组
api := engine.Group("/api")

// API v1 分组
v1 := api.Group("/v1")

// 嵌套分组示例
v1.GET("/users", handler)           // /api/v1/users
v1.GET("/users/:id", handler)        // /api/v1/users/:id
```

**中间件组合说明：**

子分组不继承父分组的中间件。`ServeHTTP` 通过遍历所有分组，按 URL 前缀匹配收集中间件，自动组合为完整链。例如请求 `/api/v1/users` 会同时匹配根分组（`""`）和 `/api` 分组的中间件。

```go
engine.Use(Recover())           // 根分组中间件，匹配所有路径
api := engine.Group("/api")
api.Use(Logger())               // /api 分组中间件

// 请求 /api/users 时的中间件链: [Recover, Logger, handler]
// 请求 /hello 时的中间件链:     [Recover, handler]
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 6. webHttp.go - HTTP 引擎

### `Engine` 结构体

Web 引擎核心，实现了 `http.Handler` 接口。

```go
type Engine struct {
	*RouterGroup              // 嵌入路由分组
	router *router            // 路由管理器
	groups  []*RouterGroup    // 所有注册的分组
}
```

| 字段 | 类型 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `RouterGroup` | `*RouterGroup` | 嵌入的路由分组，提供 `GET/POST` 方法 |
| `router` | `*router` | 内部路由管理器 |
| `groups` | `[]*RouterGroup` | 所有路由分组（用于中间件匹配）|

### 核心方法

| 方法 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `NewEngine()` | 创建新的 Engine 实例 |
| `GET(pattern, handler)` | 注册 GET 路由 |
| `POST(pattern, handler)` | 注册 POST 路由 |
| `addRoute(method, pattern, handler)` | 通用路由注册 |
| `Run(addr string)` | 启动 HTTP 服务器 |
| `ServeHTTP(w, r)` | HTTP 请求入口（实现 http.Handler）|

### 请求处理流程

```
HTTP Request
    │
    ▼
ServeHTTP(w, r)
    │
    ▼
遍历所有分组，收集匹配的中间件
    │
    ▼
newContext(w, r) ── 创建 Context
    │
    ▼
middlewares = [分组1中间件] + [分组2中间件] + [handler]
    │
    ▼
router.handle(c) ── 执行中间件链
```

### `NewEngine` 函数

创建 Engine 实例，内部将根 `RouterGroup` 加入 `groups` 列表，使 `engine.Use()` 添加的全局中间件能被 `ServeHTTP` 正确匹配。

```go
func NewEngine() *Engine {
    engine := &Engine{
        router: newRouter(),
        groups: make([]*RouterGroup, 0),
    }
    engine.RouterGroup = &RouterGroup{engine: engine}
    engine.groups = append(engine.groups, engine.RouterGroup)
    return engine
}
```

### 中间件匹配逻辑

```go
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	// ...
}
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 7. 完整使用示例

### 基础用法

```go
package main

import (
	"fmt"
	"goProject/web/http"
)

func main() {
	engine := http.NewEngine()

	// 基础路由
	engine.GET("/hello", func(c *http.Context) {
		c.String(200, "Hello %s", "World")
	})

	// 带参数的路由
	engine.GET("/user/:id", func(c *http.Context) {
		id := c.Param("id")
		c.String(200, fmt.Sprintf("User ID: %s", id))
	})

	// 启动服务器
	engine.Run(":8080")
}
```

### 使用中间件

```go
func main() {
	engine := http.NewEngine()

	// 全局中间件
	engine.Use(http.Recover())

	// 路由分组 + 分组中间件
	api := engine.Group("/api")
	api.Use(Logger())

	api.GET("/users", func(c *http.Context) {
		c.String(200, "User List")
	})
}
```

### 静态文件服务

```go
engine.GET("/static/*filepath", func(c *http.Context) {
	filepath := c.Param("filepath")
	// 使用 filepath 提供静态文件
})
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 8. 数据流图

### 8.1 请求处理流程

```
┌─────────────────────────────────────────────────────────┐
│                      HTTP Request                        │
│                   GET /api/v1/users/123                  │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                    Engine.ServeHTTP()                   │
│  1. 收集匹配的中间件                                      │
│  2. 创建 Context                                         │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                     router.handle()                      │
│  1. Trie 树查找路由                                       │
│  2. 提取路径参数 (:id = 123)                              │
│  3. 组装 handlers 链                                    │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                    Context.Next()                        │
│  handler[0] (中间件1) ──▶ handler[1] (中间件2) ──▶ ...   │
└─────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────┐
│                       Handler                            │
│  func(c *Context) {                                     │
│      id := c.Param("id")  // "123"                      │
│      c.String(200, "User ID: %s", id)                   │
│  }                                                      │
└─────────────────────────────────────────────────────────┘
```

### 8.2 路由匹配流程

```
路由注册: engine.GET("/user/:id/profile", handler)

Trie 树结构:
root
└── "user"
    └── ":id" (IsWild=true)
        └── "profile"

请求: GET /user/123/profile

匹配过程:
1. split("/user/123/profile") → ["user", "123", "profile"]
2. 查找 "user" → 找到节点
3. 查找 "123" → 匹配 ":id" (IsWild=true)
4. 查找 "profile" → 找到叶子节点
5. 提取参数: params["id"] = "123"
```

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 9. 线程安全说明

| 组件 | 线程安全 | 说明 |
|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

----|### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---|
| `Context` | 否 | 每个请求独立创建 |
| `Engine` | 部分 | 路由注册应在启动前完成 |
| `router` | 部分 | 路由注册应在启动前完成 |
| `RouterGroup` | 部分 | 分组创建应在启动前完成 |
| `trie.Node` | 否 | 路由注册应在启动前完成 |

**注意：** 当前实现主要面向单线程路由注册和多线程请求处理的场景。如需动态路由注册，请自行添加同步机制。

### `Fail` 方法

返回错误响应，直接写入状态码和消息。仅在状态码未设置时（默认 200）写入 header，避免重复写入。

```go
func (c *Context) Fail(serverError int, s string) {
    c.SetHeader("Content-Type", "text/plain")
    if c.StatusCode == 200 {
        c.Writer.WriteHeader(serverError)
    }
    _, _ = c.Writer.Write([]byte(s))
}
```

---

## 10. 与标准库的集成

GeeWeb 实现了 `http.Handler` 接口，可直接与标准库集成：

```go
// 作为 http.Server 的处理器
server := &http.Server{
	Addr:    ":8080",
	Handler: engine,
}

// 或与 mux 配合使用
mux := http.NewServeMux()
mux.Handle("/gee/", engine)
```

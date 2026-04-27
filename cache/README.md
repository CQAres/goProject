# GeeCache 分布式缓存系统

GeeCache 是一个高性能的分布式缓存系统，采用 Go 语言实现，支持 LRU-K 缓存淘汰策略、一致性哈希负载均衡和 gRPC/HTTP 双协议通信。

## 项目架构

```
cache/
├── core/           # 核心缓存逻辑
│   ├── cache.go    # 并发安全的缓存包装器
│   ├── group.go    # 缓存组管理（命名空间）
│   ├── lru-k.go    # LRU-K 缓存实现
│   └── consistenthash.go  # 一致性哈希
├── http/           # HTTP/gRPC 通信层
│   └── http.go     # HTTP 池和 peer 通信
├── protobuff/      # Protocol Buffers 定义
│   └── pb.proto    # gRPC 服务定义
└── util/           # 工具类
    └── util.go     # 通用工具和接口
```

---

## 1. core 包 - 核心缓存实现

### 1.1 cache.go

#### `cache` 结构体

并发安全的缓存包装器，封装了 LRU-K 缓存。

```go
type cache struct {
    mutex      sync.Mutex  // 互斥锁，保证并发安全
    lru        *LruCache   // LRU-K 缓存实例
    cacheBytes int64       // 最大缓存字节数
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `mutex` | `sync.Mutex` | 读写锁，保证并发安全 |
| `lru` | `*LruCache` | LRU-K 缓存实例，延迟初始化 |
| `cacheBytes` | `int64` | 缓存最大容量（字节）|

**特点：**
- 延迟初始化：LRU 缓存只在首次使用时创建
- 线程安全：所有操作通过互斥锁保护
- 简单代理：直接委托给底层的 LRU-K 实现

**方法：**

- `add(key string, k int, value util.ByteView)` - 添加缓存项
- `get(key string) (value util.ByteView, ok bool)` - 获取缓存项

---

### 1.2 group.go

#### `Group` 结构体

缓存组，提供命名空间隔离和完整的缓存生命周期管理。

```go
type Group struct {
    name      string          // 缓存组名称（命名空间）
    getter    Getter          // 缓存未命中时的数据源
    mainCache cache           // 主缓存实例
    peers     util.PeerPicker // 节点选择器（分布式支持）
    loader    *util.Group     // 请求合并器（防止缓存击穿）
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | `string` | 缓存组唯一标识，用于区分不同业务数据 |
| `getter` | `Getter` | 数据源接口，缓存未命中时调用 |
| `mainCache` | `cache` | 本地缓存实例 |
| `peers` | `util.PeerPicker` | 分布式节点选择器，可为 nil |
| `loader` | `*util.Group` | 单飞请求合并器，防止缓存击穿 |

**特点：**
- **命名空间隔离**：不同 Group 之间的 key 不会冲突
- **自动回源**：缓存未命中时自动调用 getter 获取数据
- **分布式支持**：支持从其他节点获取数据
- **防击穿**：相同 key 的并发请求会合并为一次

**核心方法：**

- `RegisterGroup(name string, cacheBytes int64, getter Getter) *Group` - 注册新缓存组
- `GetGroup(name string) *Group` - 获取已注册的缓存组
- `Get(key string) (util.ByteView, error)` - 获取缓存值
- `RegisterPeers(peers util.PeerPicker)` - 注册分布式节点选择器

**数据获取流程：**
1. 先查本地缓存，命中直接返回
2. 使用 loader 合并相同 key 的请求
3. 尝试从远程 peer 获取（如果配置了分布式）
4. 回源到 getter 获取数据
5. 写入本地缓存并返回

---

### 1.3 lru-k.go

#### `LruCache` 结构体

LRU-K 缓存实现，通过访问频率区分热点数据和冷数据。

```go
type LruCache struct {
    maxBytes      int64                         // 最大缓存容量（字节）
    nbytes        int64                         // 当前已使用字节数
    k             int                           // 晋升阈值（访问次数）
    CachelistMain *list.List                    // 主缓存链表（访问 >= k 次）
    CachelistSub  *list.List                    // 子缓存链表（访问 < k 次）
    cache         map[string]*list.Element      // key 到链表元素的映射
    history       map[string]int                // key 的访问次数统计
    OnEvicted     func(key string, value Value) // 淘汰回调函数
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `maxBytes` | `int64` | 缓存最大容量，0 表示无限制 |
| `nbytes` | `int64` | 当前缓存占用的字节数 |
| `k` | `int` | 晋升阈值，访问达到 k 次进入主缓存 |
| `CachelistMain` | `*list.List` | 主缓存链表，存储热点数据 |
| `CachelistSub` | `*list.List` | 子缓存链表，存储冷数据 |
| `cache` | `map[string]*list.Element` | 快速查找表 |
| `history` | `map[string]int` | 访问次数统计 |
| `OnEvicted` | `func` | 数据被淘汰时的回调 |

**特点：**
- **双链表结构**：区分热点和冷数据，优先淘汰冷数据
- **访问计数**：记录每个 key 的访问次数
- **自动晋升**：访问达到 k 次自动从子缓存晋升到主缓存
- **容量控制**：自动淘汰最久未使用的数据

**核心方法：**

- `NewCache(maxBytes int64, K int, onEvicted func(string, Value)) *LruCache` - 创建缓存
- `Get(key string) (Value, bool)` - 获取数据，更新访问统计
- `Add(key string, value Value)` - 添加或更新数据
- `Remove()` - 淘汰数据（优先淘汰子缓存）
- `Len() int` - 获取缓存项数量

**晋升流程：**
1. 新数据加入子缓存（CachelistSub）
2. 每次访问增加 history 计数
3. 当 history[key] == k 时，数据晋升到主缓存
4. 主缓存数据访问时移到链表头部

---

#### `entry` 结构体

缓存项的内部表示。

```go
type entry struct {
    key   string  // 缓存键
    value Value   // 缓存值（实现 Value 接口）
}
```

---

#### `Value` 接口

缓存值的抽象接口。

```go
type Value interface {
    Len() int  // 返回值占用的字节数
}
```

---

#### `Getter` 接口

数据源接口，用于缓存未命中时获取数据。

```go
type Getter interface {
    Get(key string) ([]byte, error)
}
```

---

#### `GetterFunc` 类型

函数类型，实现 Getter 接口，方便使用匿名函数。

```go
type GetterFunc func(key string) ([]byte, error)
```

---

### 1.4 consistenthash.go

#### `Map` 结构体

一致性哈希环的实现，用于分布式场景下的节点选择。

```go
type Map struct {
    hash     Hash            // 哈希函数
    replicas int             // 每个物理节点的虚拟节点数
    keys     []int           // 排序后的哈希值数组
    hashMap  map[int]string  // 哈希值到物理节点的映射
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `hash` | `Hash` | 哈希函数，默认为 crc32.ChecksumIEEE |
| `replicas` | `int` | 虚拟节点数，提高负载均衡性 |
| `keys` | `[]int` | 排序后的虚拟节点哈希值 |
| `hashMap` | `map[int]string` | 虚拟节点到物理节点的映射 |

**特点：**
- **虚拟节点**：每个物理节点对应多个虚拟节点，均匀分布
- **自动排序**：哈希环上的节点自动排序
- **顺时针查找**：key 映射到顺时针方向的第一个节点
- **均匀分布**：通过虚拟节点减少数据倾斜

**核心方法：**

- `New(replicas int, fn Hash) *Map` - 创建哈希环
- `Add(keys ...string)` - 添加物理节点
- `Get(key string) string` - 根据 key 获取对应的物理节点

---

#### `Hash` 类型

哈希函数类型定义。

```go
type Hash func(data []byte) uint32
```

---

## 2. http 包 - 通信层

### 2.1 http.go

#### `HTTPPool` 结构体

HTTP 连接池，实现 HTTP 服务器和节点间通信。

```go
type HTTPPool struct {
    self        string                // 当前节点地址
    basePath    string                // URL 前缀路径
    mu          sync.Mutex            // 互斥锁
    peers       *core.Map             // 一致性哈希环
    httpGetters map[string]*httpGetter // peer 地址到 getter 的映射
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `self` | `string` | 当前节点地址，如 "http://localhost:8001" |
| `basePath` | `string` | URL 前缀，默认为 "/cache/" |
| `mu` | `sync.Mutex` | 保护 peers 和 httpGetters |
| `peers` | `*core.Map` | 一致性哈希环，用于节点选择 |
| `httpGetters` | `map[string]*httpGetter` | peer 通信客户端缓存 |

**特点：**
- **双协议支持**：支持 HTTP 和 gRPC 通信
- **节点管理**：动态添加和选择 peer 节点
- **实现 http.Handler**：可直接作为 HTTP 服务器使用
- **实现 PeerPicker**：支持被 Group 用于节点选择

**核心方法：**

- `NewHTTPPool(self string) *HTTPPool` - 创建 HTTP 池
- `ServeHTTP(w http.ResponseWriter, r *http.Request)` - HTTP 请求处理
- `Set(peers ...string)` - 设置所有 peer 节点
- `PickPeer(key string) (util.PeerGetter, bool)` - 选择节点

---

#### `httpGetter` 结构体

Peer 节点的通信客户端，支持 HTTP 和 gRPC 协议。

```go
type httpGetter struct {
    baseURL string                      // 基础 URL
    client  protobuff.GroupCacheClient  // gRPC 客户端
    conn    *grpc.ClientConn            // gRPC 连接
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `baseURL` | `string` | 节点基础 URL |
| `client` | `protobuff.GroupCacheClient` | gRPC 客户端，可为 nil |
| `conn` | `*grpc.ClientConn` | gRPC 连接，用于关闭 |

**特点：**
- **协议自适应**：优先使用 gRPC，回退到 HTTP
- **连接复用**：gRPC 连接可复用，减少开销
- **实现 PeerGetter**：符合分布式通信接口

**核心方法：**

- `NewHTTPGetter(baseURL string) *httpGetter` - 创建 HTTP getter
- `NewGRPCGetter(target string) (*httpGetter, error)` - 创建 gRPC getter
- `Get(in *protobuff.Request, out *protobuff.Response) error` - 获取数据
- `Close() error` - 关闭连接

---

## 3. protobuff 包 - 协议定义

### 3.1 pb.proto

#### `Request` 消息

缓存请求消息。

```protobuf
message Request {
  string group = 1;  // 缓存组名称
  string key = 2;    // 缓存键
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `group` | `string` | 缓存组名称，用于命名空间隔离 |
| `key` | `string` | 缓存键 |

---

#### `Response` 消息

缓存响应消息。

```protobuf
message Response {
  bytes value = 1;   // 缓存值
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `value` | `bytes` | 缓存的二进制值 |

---

#### `GroupCache` 服务

gRPC 服务定义。

```protobuf
service GroupCache {
  rpc Get(Request) returns (Response);
}
```

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Get` | `Request` | `Response` | 获取缓存值 |

---

## 4. util 包 - 工具类

### 4.1 util.go

#### `ByteView` 结构体

不可变的字节视图，用于表示缓存值。

```go
type ByteView struct {
    B []byte  // 底层字节数组
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `B` | `[]byte` | 底层字节数据（只读）|

**特点：**
- **不可变性**：ByteSlice() 返回的是副本，保护原始数据
- **多格式支持**：支持获取字节数组和字符串形式
- **实现 Value 接口**：可用于 LRU 缓存

**方法：**

- `Len() int` - 返回字节长度
- `ByteSlice() []byte` - 返回字节副本
- `String() string` - 返回字符串形式

---

#### `PeerPicker` 接口

节点选择器接口。

```go
type PeerPicker interface {
    PickPeer(key string) (peer PeerGetter, ok bool)
}
```

| 方法 | 说明 |
|------|------|
| `PickPeer` | 根据 key 选择一个 peer 节点 |

---

#### `PeerGetter` 接口

Peer 数据获取接口。

```go
type PeerGetter interface {
    Get(in *protobuff.Request, out *protobuff.Response) error
}
```

| 方法 | 说明 |
|------|------|
| `Get` | 从 peer 获取数据 |

---

#### `Group` 结构体

请求合并器，防止缓存击穿。

```go
type Group struct {
    mu    sync.Mutex          // 互斥锁
    calls map[string]*call    // 进行中的请求映射
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `mu` | `sync.Mutex` | 保护 calls 映射 |
| `calls` | `map[string]*call` | 记录进行中的请求 |

**特点：**
- **请求合并**：相同 key 的并发请求合并为一次
- **等待机制**：后续请求等待首次请求完成，共享结果
- **自动清理**：请求完成后自动从映射中移除

**方法：**

- `Do(key string, fn func() (interface{}, error)) (interface{}, error)` - 执行或等待请求

---

#### `call` 结构体

单个请求的封装。

```go
type call struct {
    wg  sync.WaitGroup  // 等待组
    val interface{}     // 请求结果
    err error           // 请求错误
}
```

---

#### `CloneBytes` 函数

字节数组深拷贝。

```go
func CloneBytes(b []byte) []byte
```

---

## 5. 数据流图

### 5.1 本地缓存获取流程

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│ Group.Get   │────▶│  mainCache  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                                │
                       ┌────────────────────────┘
                       │ miss
                       ▼
              ┌─────────────────┐
              │   getter.Get    │
              │  (load from DB) │
              └────────┬────────┘
                       │
                       ▼
              ┌─────────────────┐
              │ populateCache   │
              └─────────────────┘
```

### 5.2 分布式缓存获取流程

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│ Group.Get   │────▶│  mainCache  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                                │
                       ┌────────────────────────┘
                       │ miss
                       ▼
              ┌─────────────────┐
              │  PickPeer       │
              │ (consistent hash)│
              └────────┬────────┘
                       │
           ┌───────────┴───────────┐
           │                       │
           ▼                       ▼
   ┌───────────────┐      ┌───────────────┐
   │  peer.Get()   │      │  getter.Get() │
   │  (remote)     │      │  (local DB)   │
   └───────────────┘      └───────────────┘
```

---

## 6. 使用示例

### 6.1 创建缓存组

```go
// 定义数据源
getter := core.GetterFunc(func(key string) ([]byte, error) {
    // 从数据库加载数据
    return db.Get(key)
})

// 注册缓存组
group := core.RegisterGroup("users", 100<<20, getter) // 100MB
```

### 6.2 使用缓存

```go
// 获取数据
view, err := group.Get("user:123")
if err != nil {
    log.Fatal(err)
}
fmt.Println(view.String())
```

### 6.3 配置分布式节点

```go
// 创建 HTTP 池
pool := http.NewHTTPPool("http://localhost:8001")

// 设置所有节点
pool.Set(
    "http://localhost:8001",
    "http://localhost:8002",
    "http://localhost:8003",
)

// 注册到缓存组
group.RegisterPeers(pool)

// 启动 HTTP 服务器
http.ListenAndServe(":8001", pool)
```

---

## 7. 性能优化点

1. **LRU-K 算法**：区分热点和冷数据，提高缓存命中率
2. **请求合并**：防止缓存击穿，减少后端压力
3. **一致性哈希**：节点增减时最小化数据迁移
4. **虚拟节点**：提高负载均衡性，减少数据倾斜
5. **gRPC 通信**：二进制协议，比 HTTP/JSON 更高效
6. **延迟初始化**：LRU 缓存按需创建，减少内存占用

---

## 8. 线程安全说明

| 组件 | 线程安全 | 机制 |
|------|----------|------|
| `cache` | 是 | `sync.Mutex` |
| `Group` | 是 | `sync.RWMutex`（groups 映射）|
| `LruCache` | 否 | 需外部加锁（由 cache 包装）|
| `HTTPPool` | 是 | `sync.Mutex` |
| `Map`（一致性哈希）| 读安全 | 写时需外部加锁 |
| `util.Group` | 是 | `sync.Mutex` |

---

## 9. 扩展点

1. **自定义 Getter**：实现任意数据源（Redis、MySQL、API 等）
2. **自定义 Hash**：替换一致性哈希的哈希函数
3. **淘汰回调**：通过 `OnEvicted` 实现缓存淘汰通知
4. **协议扩展**：实现 `PeerGetter` 支持更多通信协议

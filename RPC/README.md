# GeeRPC 分布式远程过程调用框架

GeeRPC 是一个基于 Go 语言实现的分布式 RPC 框架，支持 Gob/JSON 编解码、服务注册与发现、负载均衡、超时控制和心跳保活。

## 项目架构

```
RPC/
├── main.go              # 示例入口，演示注册中心 + 多服务端 + 客户端调用
├── client/
│   └── client.go        # RPC 客户端，支持同步/异步调用和超时控制
├── codec/
│   ├── codec.go         # 编解码器接口定义
│   └── gob.go           # Gob 编解码器实现
├── service/
│   ├── server.go        # RPC 服务端，连接管理、请求分发
│   └── service.go       # 服务方法注册与反射调用
├── xclient/
│   ├── discovery.go     # 服务发现接口与多服务器发现实现
│   └── xclient.go       # 负载均衡客户端、注册中心发现、广播调用
└── registry/
    └── registry.go      # 基于 HTTP 的服务注册中心与心跳机制
```

---

## 1. codec 包 - 编解码器

### 1.1 codec.go

#### `Header` 结构体

RPC 消息头，包含服务方法名、序列号和错误信息。

```go
type Header struct {
    ServiceMethod string  // 调用的服务方法名，格式为 "Service.Method"
    Seq           uint64  // 请求的序列号，用于匹配请求与响应
    Error         string  // 服务端返回的错误信息，空表示无错误
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `ServiceMethod` | `string` | 服务方法名，如 `"Compute.Sum"` |
| `Seq` | `uint64` | 请求序列号，客户端用于匹配请求与响应 |
| `Error` | `string` | 服务端错误信息，空字符串表示成功 |

---

#### `Codec` 接口

编解码器接口，抽象 RPC 消息的读写能力。

```go
type Codec interface {
    io.Closer
    ReadHeader(*Header) error
    ReadBody(interface{}) error
    Write(*Header, interface{}) error
}
```

| 方法 | 说明 |
|------|------|
| `ReadHeader` | 读取消息头 |
| `ReadBody` | 读取消息体 |
| `Write` | 写入消息头和消息体 |
| `Close` | 关闭底层连接 |

---

#### `NewCodecFunc` 类型与 `NewCodecFuncMap`

编解码器工厂函数类型和注册表。

```go
type NewCodecFunc func(io.ReadWriteCloser) Codec
type Type string

const (
    GobType  Type = "application/gob"
    JsonType Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc
```

**特点：**
- **工厂模式**：通过 `Type` 索引创建对应的编解码器
- **可扩展**：新增编解码器只需实现 `Codec` 接口并注册到 `NewCodecFuncMap`

---

### 1.2 gob.go

#### `GobCodec` 结构体

基于 gob 编码的 Codec 实现。

```go
type GobCodec struct {
    conn io.ReadWriteCloser  // 底层网络连接
    buf  *bufio.Writer       // 写缓冲，减少网络 IO 次数
    dec  *gob.Decoder        // gob 解码器
    enc  *gob.Encoder        // gob 编码器
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `conn` | `io.ReadWriteCloser` | 底层 TCP 连接 |
| `buf` | `*bufio.Writer` | 写缓冲，批量写入减少网络 IO |
| `dec` | `*gob.Decoder` | 从连接读取并解码 |
| `enc` | `*gob.Encoder` | 编码后写入缓冲 |

**特点：**
- **缓冲写入**：通过 `bufio.Writer` 减少网络 IO 次数
- **自动刷新**：`Write` 完成后自动 `Flush`
- **错误关闭**：编码失败时自动关闭连接

**核心方法：**

- `ReadHeader(header *Header) error` - 读取消息头
- `ReadBody(body interface{}) error` - 读取消息体
- `Write(header *Header, body interface{}) error` - 写入消息头和消息体
- `NewGobCodec(conn io.ReadWriteCloser) Codec` - 创建 GobCodec 实例

---

## 2. service 包 - 服务端

### 2.1 server.go

#### `Option` 结构体

RPC 连接协商选项，客户端在建立连接时发送。

```go
type Option struct {
    MagicNumber    int        // 魔数，验证连接合法性
    CodecType      codec.Type // 编解码器类型
    ConnectTimeOut time.Duration // 连接超时
    HandleTimeOut  time.Duration // 处理超时
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `MagicNumber` | `int` | 魔数 `0x3bef5c`，验证连接合法性 |
| `CodecType` | `codec.Type` | 编解码器类型，默认 `GobType` |
| `ConnectTimeOut` | `time.Duration` | 连接超时，默认 10 秒 |
| `HandleTimeOut` | `time.Duration` | 请求处理超时，默认 30 秒 |

---

#### `Server` 结构体

RPC 服务端实例，管理已注册的服务。

```go
type Server struct {
    serviceMap sync.Map  // 服务名到 service 实例的映射
}
```

**特点：**
- **并发安全**：使用 `sync.Map` 存储服务
- **多服务支持**：支持注册多个不同类型的服务
- **超时处理**：支持请求处理超时控制
- **全局默认**：提供 `DefaultServer` 和包级别函数

**核心方法：**

- `NewServer() *Server` - 创建新的服务端实例
- `Register(rcvr interface{}) error` - 注册服务对象
- `Accept(listener net.Listener)` - 监听并接受连接
- `ServeConn(conn io.ReadWriteCloser)` - 处理单个连接

**连接处理流程：**
1. 读取 JSON 编码的协商选项 `Option`
2. 验证 `MagicNumber` 和 `CodecType`
3. 创建对应的编解码器
4. 循环读取请求并并发处理

---

#### `request` 结构体

RPC 请求内部结构，封装消息头、参数和返回值。

```go
type request struct {
    header       *codec.Header  // 请求消息头
    argv, replyv reflect.Value  // 请求参数和响应返回值
    mtype        *methodType    // 方法类型信息
    svc          *service       // 所属服务
}
```

---

#### `handleRequest` 方法

处理单个 RPC 请求，支持超时控制。

```go
func (server *Server) handleRequest(cc codec.Codec, req *request,
    sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration)
```

**特点：**
- **超时机制**：使用 `select` + `time.After` 实现请求处理超时
- **并发安全**：通过 `sending` 互斥锁保证响应写入顺序
- **异步执行**：服务方法在独立 goroutine 中执行

---

### 2.2 service.go

#### `service` 结构体

服务对象封装，通过反射管理服务方法。

```go
type service struct {
    name   string                  // 服务名称（类型名）
    typ    reflect.Type            // 服务类型
    rcvr   reflect.Value           // 服务实例的反射值
    method map[string]*methodType  // 方法名到方法信息的映射
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | `string` | 服务名称，即结构体类型名，如 `"Compute"` |
| `typ` | `reflect.Type` | 服务类型（指针类型） |
| `rcvr` | `reflect.Value` | 服务实例，用于方法调用 |
| `method` | `map[string]*methodType` | 已注册的方法映射 |

**方法注册规则：**
- 方法必须满足签名：`func (t *T) MethodName(argType T1, replyType *T2) error`
- 参数和返回值类型必须是导出类型或内建类型
- 方法必须导出

---

#### `methodType` 结构体

方法类型信息，记录方法的反射信息。

```go
type methodType struct {
    method    reflect.Method  // 方法反射对象
    ArgType   reflect.Type    // 参数类型
    ReplyType reflect.Type    // 返回值类型（指针）
    numCalls  uint64          // 调用计数
}
```

**核心方法：**

- `newArgv() reflect.Value` - 创建参数实例，值类型自动解引用
- `newReply() reflect.Value` - 创建返回值实例，支持 Map/Slice/基本类型
- `call(m *methodType, argv, reply reflect.Value) error` - 反射调用服务方法

---

## 3. client 包 - RPC 客户端

### 3.1 client.go

#### `Call` 结构体

单个 RPC 调用的封装。

```go
type Call struct {
    Seq           uint64         // 序列号
    ServiceMethod string         // 服务方法名
    Args          interface{}    // 请求参数
    Reply         interface{}    // 响应值
    Error         error          // 错误信息
    Done          chan *Call     // 完成通知通道
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `Seq` | `uint64` | 请求序列号，与响应匹配 |
| `ServiceMethod` | `string` | 调用的服务方法名 |
| `Args` | `interface{}` | 请求参数 |
| `Reply` | `interface{}` | 响应值指针 |
| `Error` | `error` | 调用错误 |
| `Done` | `chan *Call` | 调用完成时发送通知 |

---

#### `Client` 结构体

RPC 客户端，支持同步/异步调用和超时控制。

```go
type Client struct {
    cc       codec.Codec           // 编解码器
    opt      *service.Option       // 协商选项
    sending  sync.Mutex            // 发送锁
    header   codec.Header          // 复用的消息头
    mu       sync.Mutex            // 保护 pending 和状态
    seq      uint64                // 序列号计数器
    pending  map[uint64]*Call      // 待处理请求映射
    closing  bool                  // 主动关闭标志
    shutdown bool                  // 异常关闭标志
}
```

**特点：**
- **连接超时**：`dialTimeout` 支持连接超时控制
- **异步调用**：`Go` 方法支持异步调用模式
- **同步调用**：`Call` 方法支持 `context.Context` 超时取消
- **请求-响应匹配**：通过序列号匹配请求与响应

**核心方法：**

- `Dial(network, address string, opts ...*service.Option) (*Client, error)` - 连接 RPC 服务端
- `Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call` - 异步调用
- `Call(ctx context.Context, serviceMethod string, args, reply interface{}) error` - 同步调用（支持超时）
- `Close() error` - 关闭连接
- `IsAvailable() bool` - 检查连接是否可用

**响应接收流程：**
1. 后台 goroutine 持续读取响应
2. 通过序列号从 `pending` 中找到对应的 `Call`
3. 根据响应头中的 `Error` 决定是否读取 `body`
4. 通过 `Done` 通道通知调用方

---

## 4. registry 包 - 服务注册中心

### 4.1 registry.go

#### `Registry` 结构体

基于 HTTP 的服务注册中心，管理服务端地址列表和心跳。

```go
type Registry struct {
    timeout time.Duration         // 服务超时时间
    mu      sync.Mutex            // 互斥锁
    servers map[string]*ServerItem // 服务地址到信息的映射
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `timeout` | `time.Duration` | 服务超时时间，默认 10 秒 |
| `mu` | `sync.Mutex` | 保护 servers 映射 |
| `servers` | `map[string]*ServerItem` | 已注册的服务列表 |

---

#### `ServerItem` 结构体

服务端信息，记录地址和最后心跳时间。

```go
type ServerItem struct {
    Addr  string     // 服务端地址
    start time.Time  // 最后心跳时间
}
```

**特点：**
- **自动过期**：超过 `timeout` 未心跳的服务自动从列表中移除
- **HTTP 协议**：实现 `http.Handler` 接口，通过 HTTP 提供服务
- **心跳保活**：服务端定期发送 POST 请求维持注册状态

**HTTP 接口：**

| 方法 | 路径 | Header | 说明 |
|------|------|--------|------|
| `GET` | `/registry` | 响应 `X-rpc-Registry` | 获取存活服务列表（逗号分隔） |
| `POST` | `/registry` | 请求 `X-rpc-Registry` | 注册/续期服务端地址 |

---

#### `Heartbeat` 函数

心跳保活函数，定期向注册中心发送心跳。

```go
func Heartbeat(registry, addr string, duration time.Duration)
```

| 参数 | 说明 |
|------|------|
| `registry` | 注册中心地址，如 `"http://localhost:9999/registry"` |
| `addr` | 服务端自身地址，如 `"tcp@127.0.0.1:8080"` |
| `duration` | 心跳间隔，0 则使用默认值（超时时间 - 1 秒） |

**特点：**
- **首次立即发送**：启动时立即发送一次心跳
- **定时续期**：使用 `time.Ticker` 定期发送
- **失败停止**：心跳发送失败后自动停止

---

## 5. xclient 包 - 负载均衡客户端

### 5.1 discovery.go

#### `Discovery` 接口

服务发现接口，定义服务列表的获取和更新能力。

```go
type Discovery interface {
    Refresh() error
    Update(servers []string) error
    Get(mode SelectMode) (string, error)
    GetAll() ([]string, error)
}
```

| 方法 | 说明 |
|------|------|
| `Refresh` | 从注册中心刷新服务列表 |
| `Update` | 更新本地服务列表 |
| `Get` | 根据负载均衡策略选择一个服务端 |
| `GetAll` | 获取所有可用服务端 |

---

#### `SelectMode` 类型

负载均衡策略枚举。

```go
type SelectMode int

const (
    RandomSelect     SelectMode = iota  // 随机选择
    RoundRobinSelect                     // 轮询选择
)
```

---

#### `MultiServersDiscovery` 结构体

静态多服务器发现实现，适用于服务端地址固定的场景。

```go
type MultiServersDiscovery struct {
    r       *rand.Rand       // 随机数生成器
    mu      sync.RWMutex     // 读写锁
    servers []string         // 服务端地址列表
    index   int              // 轮询索引
}
```

**特点：**
- **随机选择**：从列表中随机选取一个服务端
- **轮询选择**：按顺序依次选择，循环往复
- **线程安全**：通过读写锁保护并发访问

---

### 5.2 xclient.go

#### `XClient` 结构体

支持负载均衡和连接复用的 RPC 客户端。

```go
type XClient struct {
    d       Discovery                  // 服务发现
    mode    SelectMode                 // 负载均衡策略
    opt     *service.Option            // 协商选项
    mu      sync.Mutex                 // 保护 clients
    clients map[string]*client2.Client // 地址到客户端的映射（连接复用）
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `d` | `Discovery` | 服务发现实例 |
| `mode` | `SelectMode` | 负载均衡策略 |
| `opt` | `*service.Option` | 连接选项 |
| `clients` | `map[string]*client2.Client` | 已建立的客户端连接缓存 |

**特点：**
- **连接复用**：相同地址的客户端连接会被缓存和复用
- **自动重连**：检测到不可用连接时自动关闭并重新建立
- **地址解析**：自动去除 `"tcp@"` 前缀

**核心方法：**

- `Call(ctx context.Context, serviceMethod string, args, reply interface{}) error` - 单点调用
- `Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error` - 广播调用

---

#### `RegistryDiscovery` 结构体

基于注册中心的服务发现实现，定期从注册中心刷新服务列表。

```go
type RegistryDiscovery struct {
    *MultiServersDiscovery              // 嵌入静态发现
    registry   string                   // 注册中心地址
    timeout    time.Duration            // 刷新间隔
    lastUpdate time.Time                // 上次刷新时间
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `registry` | `string` | 注册中心 HTTP 地址 |
| `timeout` | `time.Duration` | 刷新间隔，默认 10 秒 |
| `lastUpdate` | `time.Time` | 上次刷新时间 |

**特点：**
- **懒刷新**：超过 `timeout` 才触发刷新，避免频繁请求
- **HTTP 拉取**：通过 `GET /registry` 获取存活服务列表
- **继承静态发现**：复用随机和轮询选择逻辑

---

#### `Broadcast` 方法

向所有服务端广播调用，任一成功即返回。

```go
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string,
    args, reply interface{}) error
```

**特点：**
- **并发调用**：同时向所有服务端发起请求
- **快速失败**：任一请求失败且无成功时，取消其余请求
- **首次成功**：第一个成功的响应写入 reply
- **上下文取消**：支持通过 `context.WithCancel` 取消所有进行中的请求

---

## 6. 数据流图

### 6.1 单次调用流程

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Client  │────▶│   Dial   │────▶│  Option  │────▶│  Server  │
│  Call()  │     │  TCP     │     │  JSON    │     │ServeConn │
└──────────┘     └──────────┘     └──────────┘     └────┬─────┘
                                                         │
                       ┌─────────────────────────────────┘
                       │ 验证 MagicNumber + CodecType
                       ▼
              ┌───────────────┐     ┌───────────────┐
              │  serveCodec   │────▶│ readRequest   │
              │  循环读取请求  │     │ Header+Body   │
              └───────┬───────┘     └───────┬───────┘
                      │                     │
                      ▼                     ▼
              ┌───────────────┐     ┌───────────────┐
              │ handleRequest │────▶│  svc.call()   │
              │  超时控制      │     │  反射调用方法  │
              └───────┬───────┘     └───────┬───────┘
                      │                     │
                      ▼                     ▼
              ┌───────────────┐     ┌───────────────┐
              │ sendResponse  │────▶│ Client.receive│
              │ Header+Body   │     │ 匹配Seq+通知  │
              └───────────────┘     └───────────────┘
```

### 6.2 服务注册与发现流程

```
┌──────────┐  POST /registry  ┌──────────┐  GET /registry  ┌──────────┐
│  Server  │────────────────▶│ Registry │────────────────▶│ XClient  │
│ Heartbeat│  X-rpc-Registry │  HTTP    │  X-rpc-Registry │ Discovery│
└──────────┘  tcp@addr:port  └──────────┘  addr1,addr2    └──────────┘
      │                            │                            │
      │  定期心跳(9s)              │  过期淘汰(10s)            │  定期刷新(10s)
      └────────────────────────────┴────────────────────────────┘
```

### 6.3 广播调用流程

```
                    ┌──────────┐
                    │ XClient  │
                    │Broadcast │
                    └────┬─────┘
                         │
           ┌─────────────┼─────────────┐
           ▼             ▼             ▼
    ┌─────────────┐┌─────────────┐┌─────────────┐
    │  Server 1   ││  Server 2   ││  Server 3   │
    │  call()     ││  call()     ││  call()     │
    └──────┬──────┘└──────┬──────┘└──────┬──────┘
           │              │              │
           │   首次成功写入 reply        │
           ▼              ▼              ▼
    ┌─────────────────────────────────────────┐
    │          WaitGroup.Wait()               │
    │          返回第一个错误（如有）            │
    └─────────────────────────────────────────┘
```

---

## 7. 使用示例

### 7.1 基础用法（单服务端）

```go
// 定义服务
type Compute int

func (c *Compute) Sum(args Args, reply *int) error {
    *reply = args.Num1 + args.Num2
    return nil
}

// 启动服务端
func startServer(addr chan string) {
    var c Compute
    service.Register(&c)
    l, _ := net.Listen("tcp", ":0")
    addr <- l.Addr().String()
    service.Accept(l)
}

// 客户端调用
client, _ := client.Dial("tcp", <-addr)
defer client.Close()
var reply int
err := client.Call(context.Background(), "Compute.Sum", &Args{1, 2}, &reply)
```

### 7.2 分布式用法（注册中心 + 多服务端）

```go
// 启动注册中心
registry.HandleHTTP()
http.ListenAndServe(":9999", nil)

// 启动服务端（自动心跳注册）
l, _ := net.Listen("tcp", ":0")
server := service.NewServer()
server.Register(&Compute{})
registry.Heartbeat("http://localhost:9999/registry", "tcp@"+l.Addr().String(), 0)
server.Accept(l)

// 客户端通过注册中心发现服务
d := xclient.NewRegistryDiscovery("http://localhost:9999/registry", 0)
xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
defer xc.Close()
var reply int
xc.Call(context.Background(), "Compute.Sum", &xclient.Args{1, 2}, &reply)

// 广播调用所有服务端
xc.Broadcast(context.Background(), "Compute.Sum", &xclient.Args{1, 2}, &reply)
```

---

## 8. 通信协议

### 8.1 连接协商阶段

```
Client                          Server
  │                               │
  │──── JSON(Option) ────────────▶│  MagicNumber + CodecType
  │                               │  验证合法性
  │◀─── 后续使用协商的编解码器 ────│
```

### 8.2 请求/响应阶段（Gob 编码）

```
Client                          Server
  │                               │
  │──── Gob(Header) ─────────────▶│  ServiceMethod + Seq + Error
  │──── Gob(Body)   ─────────────▶│  请求参数
  │                               │
  │◀─── Gob(Header) ─────────────│  ServiceMethod + Seq + Error
  │◀─── Gob(Body)   ─────────────│  响应结果
```

---

## 9. 线程安全说明

| 组件 | 线程安全 | 机制 |
|------|----------|------|
| `Server` | 是 | `sync.Map` 存储服务 |
| `service` | 部分 | 服务注册在启动前完成 |
| `Client` | 是 | `sync.Mutex` 保护发送和 pending |
| `Registry` | 是 | `sync.Mutex` 保护 servers |
| `MultiServersDiscovery` | 是 | `sync.RWMutex` 保护 servers |
| `RegistryDiscovery` | 是 | 继承 `sync.RWMutex` |
| `XClient` | 是 | `sync.Mutex` 保护 clients |

---

## 10. 扩展点

1. **自定义编解码器**：实现 `Codec` 接口，注册到 `NewCodecFuncMap`（如 JSON 编解码器）
2. **自定义服务发现**：实现 `Discovery` 接口（如基于 etcd/consul 的发现）
3. **自定义负载均衡**：扩展 `SelectMode` 枚举和 `Get` 方法
4. **自定义注册中心**：替换 `Registry`，支持分布式注册存储
5. **拦截器/中间件**：在 `handleRequest` 前后添加日志、鉴权等逻辑

package http

import (
	"context"
	"fmt"
	"goProject/cache/core"
	"goProject/cache/protobuff"
	"goProject/cache/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultBasePath = "/cache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex
	peers       *core.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := core.GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(view.ByteSlice())
	if err != nil {
		return
	}
}

type httpGetter struct {
	baseURL string
	// 用于 gRPC 通信的客户端
	client protobuff.GroupCacheClient
	conn   *grpc.ClientConn
}

// NewHTTPGetter 创建支持 gRPC 的 httpGetter
func NewHTTPGetter(baseURL string) *httpGetter {
	return &httpGetter{baseURL: baseURL}
}

// NewGRPCGetter 创建 gRPC 协议的 getter
func NewGRPCGetter(target string) (*httpGetter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", target, err)
	}

	return &httpGetter{
		baseURL: target,
		client:  protobuff.NewGroupCacheClient(conn),
		conn:    conn,
	}, nil
}

// Close 关闭 gRPC 连接
func (h *httpGetter) Close() error {
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}

// Get 实现 PeerGetter 接口，优先使用 gRPC，回退到 HTTP
func (h *httpGetter) Get(in *protobuff.Request, out *protobuff.Response) error {
	// 如果配置了 gRPC 客户端，使用 gRPC
	if h.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		resp, err := h.client.Get(ctx, in)
		if err != nil {
			return err
		}
		out.Value = resp.Value
		return nil
	}

	// 回退到 HTTP 协议
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(in.Group), url.QueryEscape(in.Key))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			return
		}
	}(res.Body)

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	out.Value = bytes
	return nil
}

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = core.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = NewHTTPGetter(peer + p.basePath)
	}
}

func (p *HTTPPool) PickPeer(key string) (util.PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

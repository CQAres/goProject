package xclient

import (
	client2 "goProject/RPC/client"
	"goProject/RPC/service"
	"io"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *service.Option
	mu      sync.Mutex
	clients map[string]*client2.Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *service.Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*client2.Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

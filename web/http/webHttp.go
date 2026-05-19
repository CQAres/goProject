package http

import (
	"net/http"
	"strings"
)

// Engine is the uni handler for all requests
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

func NewEngine() *Engine {
	engine := &Engine{
		router: newRouter(),
		groups: make([]*RouterGroup, 0),
	}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = append(engine.groups, engine.RouterGroup)
	return engine
}

func (e *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	e.router.addRoute(method, pattern, handler)
}

func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRoute("POST", pattern, handler)
}

func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

// Go 的 net/http 包中，只要你的类型实现了 http.Handler 接口 （该接口只有一个方法 ServeHTTP(w http.ResponseWriter, r *http.Request)） 就可以将它作为 HTTP 服务器的处理器来拦截所有请求
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	e.router.handle(c)
}

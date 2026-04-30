package http

import (
	"fmt"
	"net/http"
)

// Engine is the uni handler for all requests
type Engine struct {
	router *router
}

func NewEngine() *Engine {
	return &Engine{
		router: newRouter(),
	}
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

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.router.handle(c)
}

//	Go 的 net/http 包中，只要你的类型实现了 http.Handler 接口 （该接口只有一个方法 ServeHTTP(w http.ResponseWriter, r *http.Request)） 就可以将它作为 HTTP 服务器的处理器来拦截所有请求

func main() {
	engine := NewEngine()
	err := engine.Run("9999")
	if err != nil {
		fmt.Printf("run err: %v\n", err)
		return
	}
}

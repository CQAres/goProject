package http

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(*Context)

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string
	Method     string
	Params     map[string]string
	StatusCode int
	handlers   []HandlerFunc // 中间件
	index      int
}

func (c *Context) params(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Param(key string) string {
	return c.params(key)
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer:     w,
		Req:        req,
		Path:       req.URL.Path,
		Method:     req.Method,
		StatusCode: 200,
		index:      -1,
	}
}

func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.PostForm.Get(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	if _, err := c.Writer.Write([]byte(fmt.Sprintf(format, values...))); err != nil {
		return
	}
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if _, err := c.Writer.Write([]byte(html)); err != nil {
		return
	}
}

func (c *Context) Fail(serverError int, s string) {
	c.SetHeader("Content-Type", "text/plain")
	// Only write header if not already committed
	if c.StatusCode == 200 {
		c.Writer.WriteHeader(serverError)
	}
	_, _ = c.Writer.Write([]byte(s))
}

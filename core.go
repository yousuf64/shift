package dune

import "net/http"

type routeLog struct {
	method  string
	path    string
	handler HandlerFunc
}

type Core struct {
	base string
	logs *[]routeLog
	mws  []MiddlewareFunc
}

func (c *Core) Group(path string, f func(d *Group)) {
	stack := make([]MiddlewareFunc, len(c.mws), len(c.mws))
	copy(stack, c.mws)

	f(&Group{Core{
		logs: c.logs,
		base: c.base + path,
		mws:  stack,
	}})
}

func (c *Core) With(middlewares ...MiddlewareFunc) *Core {
	stack := make([]MiddlewareFunc, len(c.mws), len(c.mws)+len(middlewares))
	copy(stack, c.mws)
	stack = append(stack, middlewares...)

	return &Core{
		c.base,
		c.logs,
		stack,
	}
}

func (c *Core) Mount(path string, dune *Router) {
	for _, log := range *dune.logs {
		c.Map([]string{log.method}, path+log.path, log.handler)
	}
}

func (c *Core) Map(methods []string, path string, handler HandlerFunc) {
	if len(methods) == 0 {
		panic("methods cannot be empty")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	for _, meth := range methods {
		*c.logs = append(*c.logs, routeLog{
			method:  meth,
			path:    c.base + path,
			handler: c.chain(handler),
		})
	}
}

func (c *Core) GET(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodGet}, path, handler)
}

func (c *Core) POST(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodPost}, path, handler)
}

func (c *Core) PUT(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodPut}, path, handler)
}

func (c *Core) PATCH(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodPatch}, path, handler)
}

func (c *Core) DELETE(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodDelete}, path, handler)
}

func (c *Core) OPTIONS(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodOptions}, path, handler)
}

func (c *Core) HEAD(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodHead}, path, handler)
}

func (c *Core) CONNECT(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodConnect}, path, handler)
}

func (c *Core) TRACE(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodTrace}, path, handler)
}

// Any registers the route for all the built-in http methods and registered custom http methods.
func (c *Core) Any(path string, handler HandlerFunc) {
	c.Map([]string{""}, path, handler)
}

func (c *Core) chain(handler HandlerFunc) HandlerFunc {
	for i := len(c.mws) - 1; i >= 0; i-- {
		handler = c.mws[i](handler)
	}
	return handler
}

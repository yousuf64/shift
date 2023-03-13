package shift

import "net/http"

type routeLog struct {
	method  string
	path    string
	handler HandlerFunc
}

// Core provides methods to register routes.
type Core struct {
	base string
	logs *[]routeLog
	mws  []MiddlewareFunc
}

// Group groups routes together at the given path with a group-scoped middleware stack inherited from the parent middleware stack.
// It provides the opportunity to maintain groups of routes in different files using the func(g *Group) func signature.
//
// It is also possible to nest groups within groups.
func (c *Core) Group(path string, f func(g *Group)) {
	stack := make([]MiddlewareFunc, len(c.mws), len(c.mws))
	copy(stack, c.mws)

	f(&Group{Core{
		logs: c.logs,
		base: c.base + path,
		mws:  stack,
	}})
}

// With returns an instance attaching middlewares to the middleware stack inherited from the parent middleware stack.
// It's useful for registering middlewares for a specific Group or a route.
// To use a net/http idiomatic middleware, wrap the middleware using the HTTPMiddlewareFunc.
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

// Map maps a request handler for the given methods at the given path.
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

// GET maps a request handler for the GET method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodGet}, path, handler)
func (c *Core) GET(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodGet}, path, handler)
}

// POST maps a request handler for the POST method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodPost}, path, handler)
func (c *Core) POST(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodPost}, path, handler)
}

// PUT maps a request handler for the PUT method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodPut}, path, handler)
func (c *Core) PUT(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodPut}, path, handler)
}

// PATCH maps a request handler for the PATCH method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodPatch}, path, handler)
func (c *Core) PATCH(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodPatch}, path, handler)
}

// DELETE maps a request handler for the DELETE method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodDelete}, path, handler)
func (c *Core) DELETE(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodDelete}, path, handler)
}

// OPTIONS maps a request handler for the OPTIONS method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodOptions}, path, handler)
func (c *Core) OPTIONS(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodOptions}, path, handler)
}

// HEAD maps a request handler for the HEAD method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodHead}, path, handler)
func (c *Core) HEAD(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodHead}, path, handler)
}

// CONNECT maps a request handler for the CONNECT method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodConnect}, path, handler)
func (c *Core) CONNECT(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodConnect}, path, handler)
}

// TRACE maps a request handler for the TRACE method at the given path.
// It is a shorthand for:
//
//	c.Map([]string{http.MethodTrace}, path, handler)
func (c *Core) TRACE(path string, handler HandlerFunc) {
	c.Map([]string{http.MethodTrace}, path, handler)
}

// All maps a request handler for all the built-in HTTP methods and registered custom HTTP methods at the given path.
// It is a shorthand for:
//
//	c.Map([]string{""}, path, handler)
func (c *Core) All(path string, handler HandlerFunc) {
	c.Map([]string{""}, path, handler)
}

func (c *Core) chain(handler HandlerFunc) HandlerFunc {
	for i := len(c.mws) - 1; i >= 0; i-- {
		handler = c.mws[i](handler)
	}
	return handler
}

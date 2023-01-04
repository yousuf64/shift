package dune

import (
	"context"
	"net/http"
	"sync"
)

type MiddlewareFunc func(next Handler) Handler

type ctxKey struct{}

func HandlerFunc(handler http.Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request, p *Params) {
		r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, p))
		handler.ServeHTTP(w, r)
	}
}

type Methods []string

const (
	MethodGet     = http.MethodGet
	MethodPost    = http.MethodPost
	MethodPut     = http.MethodPut
	MethodPatch   = http.MethodPatch
	MethodDelete  = http.MethodDelete
	MethodOptions = http.MethodOptions
	MethodHead    = http.MethodHead
	MethodConnect = http.MethodConnect
	MethodTrace   = http.MethodTrace
)

type log struct {
	method  string
	path    string
	handler Handler
}

type mux struct {
	trees      map[string]*node
	paramsPool sync.Pool
	maxParams  int
	logs       []log
}

type Option interface {
	apply(router *Router)
}

type optionFunc func(*Router)

func (r optionFunc) apply(router *Router) {
	r(router)
}

func Use(middlewares ...MiddlewareFunc) Option {
	return optionFunc(func(router *Router) {
		router.mws = append(router.mws, middlewares...)
	})
}

type Router struct {
	*mux

	base string
	mws  []MiddlewareFunc
}

func New(opts ...Option) *Router {
	r := &Router{
		mux: &mux{
			trees:      map[string]*node{},
			paramsPool: sync.Pool{},
			maxParams:  0,
		},
	}

	for _, opt := range opts {
		opt.apply(r)
	}

	return r
}

func (r *Router) Group(path string, f func(r *Router)) {
	stack := make([]MiddlewareFunc, 0, len(r.mws))
	copy(stack, r.mws)

	f(&Router{
		mux:  r.mux,
		base: r.base + path,
		mws:  stack,
	})
}

func (r *Router) With(middlewares ...MiddlewareFunc) *Router {
	stack := make([]MiddlewareFunc, 0, len(r.mws)+len(middlewares))
	copy(stack, r.mws)
	stack = append(stack, middlewares...)

	return &Router{r.mux, r.base, stack}
}

func (r *Router) Mount(path string, router *Router) {
	for _, log := range router.logs {
		r.Map(Methods{log.method}, r.base+path+log.path, log.handler)
	}
}

func (r *Router) Map(methods Methods, path string, handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	for _, meth := range methods {
		tr, ok := r.mux.trees[meth]
		if !ok {
			tr = newRootNode()
			r.mux.trees[meth] = tr
		}

		pc := tr.insert(r.base+path, r.chain(handler))
		if r.paramsPool.New == nil || pc > r.maxParams {
			r.maxParams = pc
			r.paramsPool.New = func() any {
				return newParams(pc)
			}
		}

		r.mux.logs = append(r.mux.logs, log{
			method:  meth,
			path:    path,
			handler: handler,
		})
	}
}

func (r *Router) Get(path string, handler Handler) {
	r.Map(Methods{MethodGet}, path, handler)
}

func (r *Router) Post(path string, handler Handler) {
	r.Map(Methods{MethodPost}, path, handler)
}

func (r *Router) Put(path string, handler Handler) {
	r.Map(Methods{MethodPut}, path, handler)
}

func (r *Router) Patch(path string, handler Handler) {
	r.Map(Methods{MethodPatch}, path, handler)
}

func (r *Router) Delete(path string, handler Handler) {
	r.Map(Methods{MethodDelete}, path, handler)
}

func (r *Router) Options(path string, handler Handler) {
	r.Map(Methods{MethodOptions}, path, handler)
}

func (r *Router) Head(path string, handler Handler) {
	r.Map(Methods{MethodHead}, path, handler)
}

func (r *Router) Connect(path string, handler Handler) {
	r.Map(Methods{MethodConnect}, path, handler)
}

func (r *Router) Trace(path string, handler Handler) {
	r.Map(Methods{MethodTrace}, path, handler)
}

func (r *Router) Any(path string, handler Handler) {
	//TODO implement me
	panic("implement me")
}

func (r *Router) chain(handler Handler) Handler {
	for i := len(r.mws) - 1; i >= 0; i-- {
		handler = r.mws[i](handler)
	}
	return handler
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	tr, ok := r.trees[request.Method]
	if !ok {
		panic("implement method not allowed")
	}

	n, ps := tr.search(request.URL.Path, func() *Params {
		ps := r.paramsPool.Get().(*Params)
		ps.reset()
		return ps
	})

	if n != nil && n.handler != nil {
		n.handler(writer, request, ps)
		if ps != nil {
			r.paramsPool.Put(ps)
		}
		return
	}

	if ps != nil {
		ps.reset()
		r.paramsPool.Put(ps)
	}

	panic("implement method not allowed")
}

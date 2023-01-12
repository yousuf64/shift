package dune

import (
	"net/http"
	"strings"
	"sync"
)

type Action uint8

const (
	DoNone Action = iota
	DoRedirect
	DoExecute
)

type Config struct {
	OnTrailingSlashMatch   Action
	OnFixedPathMatch       Action
	NotFoundHandler        Handler
	HandleMethodNotAllowed bool
}

var defaultConfig = &Config{
	HandleMethodNotAllowed: true,
	OnTrailingSlashMatch:   DoRedirect,
	OnFixedPathMatch:       DoRedirect,
	NotFoundHandler:        defaultNotFoundHandler,
}

type MiddlewareFunc func(next Handler) Handler

type HTTPMiddlewareFunc func(next http.Handler) http.Handler

func HandlerFunc(handler http.HandlerFunc) Handler {
	return func(w http.ResponseWriter, r *http.Request, p *Params) {
		if p != nil {
			if !hasParamsCtx(r.Context()) {
				r = r.WithContext(withParamsCtx(r.Context(), p))
			}
		}
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

var methodAll = Methods{MethodGet, MethodPost, MethodPut, MethodPatch, MethodDelete, MethodOptions, MethodHead, MethodConnect, MethodTrace}

type Option interface {
	apply(d *Dune)
}

type optionFunc func(*Dune)

func (r optionFunc) apply(d *Dune) {
	r(d)
}

func Use(middlewares ...MiddlewareFunc) Option {
	return optionFunc(func(d *Dune) {
		d.mws = append(d.mws, middlewares...)
	})
}

func UseHTTP(middlewares ...HTTPMiddlewareFunc) Option {
	return optionFunc(func(d *Dune) {
		for _, mw := range middlewares {
			d.mws = append(d.mws, wrapHttpMiddleware(mw))
		}
	})
}

func OnTrailingSlashMatch(action Action) Option {
	return optionFunc(func(d *Dune) {
		d.config.OnTrailingSlashMatch = action
	})
}

func OnFixedPathMatch(action Action) Option {
	return optionFunc(func(d *Dune) {
		d.config.OnFixedPathMatch = action
	})
}

func WithNotFoundHandler(handler Handler) Option {
	return optionFunc(func(d *Dune) {
		d.config.NotFoundHandler = handler
	})
}

func SetHandleMethodNotAllowed(b bool) Option {
	return optionFunc(func(d *Dune) {
		d.config.HandleMethodNotAllowed = b
	})
}

func wrapHttpMiddleware(mw HTTPMiddlewareFunc) MiddlewareFunc {
	return func(next Handler) Handler {
		return func(w http.ResponseWriter, r *http.Request, ps *Params) {
			nextFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next(w, r, ps)
			})

			if ps != nil {
				if !hasParamsCtx(r.Context()) {
					r = r.WithContext(withParamsCtx(r.Context(), ps))
				}
			}
			mw(nextFn).ServeHTTP(w, r)
		}
	}
}

func defaultNotFoundHandler(w http.ResponseWriter, r *http.Request, _ *Params) {
	http.NotFound(w, r)
}

type log struct {
	method  string
	path    string
	handler Handler
}

type Dune struct {
	logs   *[]log
	base   string
	mws    []MiddlewareFunc
	config *Config
}

func New(opts ...Option) *Dune {
	d := &Dune{
		logs:   &[]log{},
		base:   "",
		mws:    nil,
		config: defaultConfig,
	}

	for _, opt := range opts {
		opt.apply(d)
	}

	return d
}

func (d *Dune) Group(path string, f func(d *Dune)) {
	stack := make([]MiddlewareFunc, len(d.mws), len(d.mws))
	copy(stack, d.mws)

	f(&Dune{d.logs, d.base + path, stack, d.config})
}

func (d *Dune) With(middlewares ...MiddlewareFunc) *Dune {
	stack := make([]MiddlewareFunc, len(d.mws), len(d.mws)+len(middlewares))
	copy(stack, d.mws)
	stack = append(stack, middlewares...)

	return &Dune{d.logs, d.base, stack, d.config}
}

func (d *Dune) Mount(path string, dune *Dune) {
	for _, log := range *dune.logs {
		d.Map(Methods{log.method}, path+log.path, log.handler)
	}
}

func (d *Dune) Map(methods Methods, path string, handler Handler) {
	if len(methods) == 0 {
		panic("methods cannot be empty")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	for _, meth := range methods {
		*d.logs = append(*d.logs, log{
			method:  meth,
			path:    d.base + path,
			handler: d.chain(handler),
		})
	}
}

func (d *Dune) Get(path string, handler Handler) {
	d.Map(Methods{MethodGet}, path, handler)
}

func (d *Dune) Post(path string, handler Handler) {
	d.Map(Methods{MethodPost}, path, handler)
}

func (d *Dune) Put(path string, handler Handler) {
	d.Map(Methods{MethodPut}, path, handler)
}

func (d *Dune) Patch(path string, handler Handler) {
	d.Map(Methods{MethodPatch}, path, handler)
}

func (d *Dune) Delete(path string, handler Handler) {
	d.Map(Methods{MethodDelete}, path, handler)
}

func (d *Dune) Options(path string, handler Handler) {
	d.Map(Methods{MethodOptions}, path, handler)
}

func (d *Dune) Head(path string, handler Handler) {
	d.Map(Methods{MethodHead}, path, handler)
}

func (d *Dune) Connect(path string, handler Handler) {
	d.Map(Methods{MethodConnect}, path, handler)
}

func (d *Dune) Trace(path string, handler Handler) {
	d.Map(Methods{MethodTrace}, path, handler)
}

func (d *Dune) Any(path string, handler Handler) {
	d.Map(methodAll, path, handler)
}

func (d *Dune) chain(handler Handler) Handler {
	for i := len(d.mws) - 1; i >= 0; i-- {
		handler = d.mws[i](handler)
	}
	return handler
}

// Router

type Router struct {
	multiplexers map[string]muxInterface
	config       *Config
}

func Compile(d *Dune) *Router {
	type info struct {
		static int
		wc     int
		logs   []log
	}

	r := &Router{map[string]muxInterface{}, d.config}

	methodsInfo := make(map[string]*info)

	for _, lg := range *d.logs {
		if _, ok := methodsInfo[lg.method]; !ok {
			methodsInfo[lg.method] = &info{}
		}

		inf := methodsInfo[lg.method]
		inf.logs = append(inf.logs, lg)

		static := isStatic(lg.path)

		if static {
			inf.static++
		} else {
			inf.wc++
		}
	}

	for meth, inf := range methodsInfo {
		var mux muxInterface

		total := len(inf.logs)
		staticPercentage := float64(inf.static) / float64(total) * 100

		if staticPercentage == 100 {
			mux = newStaticMux()
		} else if staticPercentage >= 30 {
			mux = newHybridMux()
		} else {
			mux = newRadixMux()
		}

		for _, log := range inf.logs {
			mux.add(log.path, log.handler)
		}

		r.multiplexers[meth] = mux
	}

	return r
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := router.multiplexers[r.Method]
	if mux == nil {
		http.NotFound(w, r)
		return
	}

	handler, ps := mux.find(r.URL.Path)
	if handler == nil {
		http.NotFound(w, r)
		return
	}

	handler(w, r, ps)
}

// Mux

type muxInterface interface {
	add(path string, handler Handler)
	find(path string) (Handler, *Params)
}

type radixMux struct {
	tree       *node
	paramsPool sync.Pool
	maxParams  int
}

func newRadixMux() *radixMux {
	return &radixMux{
		tree:       newRootNode(),
		paramsPool: sync.Pool{},
		maxParams:  0,
	}
}

func (mux *radixMux) add(path string, handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	// Wrap handler to put Params obj back to the pool after handler execution.
	pc := mux.tree.insert(path, func(w http.ResponseWriter, r *http.Request, p *Params) {
		//r = r.WithContext(context.WithValue(r.Context(), 0, p))
		handler(w, r, p)

		if p != nil {
			mux.paramsPool.Put(p)
		}
	})
	if mux.paramsPool.New == nil || pc > mux.maxParams {
		mux.maxParams = pc
		mux.paramsPool.New = func() any {
			return newParams(pc)
		}
	}
}

func (mux *radixMux) find(path string) (Handler, *Params) {
	n, ps := mux.tree.search(path, func() *Params {
		ps := mux.paramsPool.Get().(*Params)
		ps.reset()
		return ps
	})

	if n != nil && n.handler != nil {
		return n.handler, ps
	}

	if ps != nil {
		mux.paramsPool.Put(ps)
	}
	return nil, nil
}

type staticMux struct {
	routes   map[string]Handler
	sizePlot []bool
}

func newStaticMux() *staticMux {
	return &staticMux{
		routes:   map[string]Handler{},
		sizePlot: make([]bool, 5),
	}
}

func (mux *staticMux) add(path string, handler Handler) {
	if len(path) >= len(mux.sizePlot) {
		// Grow slice.
		mux.sizePlot = append(mux.sizePlot, make([]bool, len(path)-len(mux.sizePlot)+1)...)
	}

	mux.routes[path] = handler
	mux.sizePlot[len(path)] = true
}

func (mux *staticMux) find(path string) (Handler, *Params) {
	if len(path) >= len(mux.sizePlot) {
		return nil, nil
	}

	if !mux.sizePlot[len(path)] {
		return nil, nil
	}

	return mux.routes[path], nil
}

type hybridMux struct {
	static *staticMux
	radix  *radixMux
}

func newHybridMux() *hybridMux {
	return &hybridMux{newStaticMux(), newRadixMux()}
}

func (mux *hybridMux) add(path string, handler Handler) {
	static := isStatic(path)
	if static {
		mux.static.add(path, handler)
	} else {
		mux.radix.add(path, handler)
	}
}

func (mux *hybridMux) find(path string) (Handler, *Params) {
	if handler, ps := mux.static.find(path); handler != nil {
		return handler, ps
	}

	return mux.radix.find(path)
}

func isStatic(path string) bool {
	return strings.IndexFunc(path, func(r rune) bool {
		return r == ':' || r == '*'
	}) == -1
}

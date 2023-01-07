package dune

import (
	"net/http"
	"strings"
	"sync"
)

type DuneOption interface {
	apply(router *Dune)
}

type duneOptionFunc func(*Dune)

func (r duneOptionFunc) apply(router *Dune) {
	r(router)
}

func Use2(middlewares ...MiddlewareFunc) DuneOption {
	return duneOptionFunc(func(dune *Dune) {
		dune.mws = append(dune.mws, middlewares...)
	})
}

type Dune struct {
	logs *[]log
	base string
	mws  []MiddlewareFunc
}

func NewDune(opts ...DuneOption) *Dune {
	d := &Dune{
		logs: &[]log{},
		base: "",
		mws:  nil,
	}

	for _, opt := range opts {
		opt.apply(d)
	}

	return d
}

func (r *Dune) Group(path string, f func(r *Dune)) {
	stack := make([]MiddlewareFunc, len(r.mws), len(r.mws))
	copy(stack, r.mws)

	f(&Dune{
		logs: r.logs,
		base: r.base + path,
		mws:  stack,
	})
}

func (r *Dune) With(middlewares ...MiddlewareFunc) *Dune {
	stack := make([]MiddlewareFunc, len(r.mws), len(r.mws)+len(middlewares))
	copy(stack, r.mws)
	stack = append(stack, middlewares...)

	return &Dune{r.logs, r.base, stack}
}

func (r *Dune) Mount(path string, dune *Dune) {
	for _, log := range *dune.logs {
		r.Map(Methods{log.method}, r.base+path+log.path, log.handler)
	}
}

func (r *Dune) Map(methods Methods, path string, handler Handler) {
	if handler == nil {
		panic("handler cannot be nil")
	}

	for _, meth := range methods {
		*r.logs = append(*r.logs, log{
			method:  meth,
			path:    path,
			handler: r.chain(handler),
		})
	}
}

func (r *Dune) Get(path string, handler Handler) {
	r.Map(Methods{MethodGet}, path, handler)
}

func (r *Dune) Post(path string, handler Handler) {
	r.Map(Methods{MethodPost}, path, handler)
}

func (r *Dune) Put(path string, handler Handler) {
	r.Map(Methods{MethodPut}, path, handler)
}

func (r *Dune) Patch(path string, handler Handler) {
	r.Map(Methods{MethodPatch}, path, handler)
}

func (r *Dune) Delete(path string, handler Handler) {
	r.Map(Methods{MethodDelete}, path, handler)
}

func (r *Dune) Options(path string, handler Handler) {
	r.Map(Methods{MethodOptions}, path, handler)
}

func (r *Dune) Head(path string, handler Handler) {
	r.Map(Methods{MethodHead}, path, handler)
}

func (r *Dune) Connect(path string, handler Handler) {
	r.Map(Methods{MethodConnect}, path, handler)
}

func (r *Dune) Trace(path string, handler Handler) {
	r.Map(Methods{MethodTrace}, path, handler)
}

func (r *Dune) Any(path string, handler Handler) {
	r.Map(methodAll, path, handler)
}

func (r *Dune) chain(handler Handler) Handler {
	for i := len(r.mws) - 1; i >= 0; i-- {
		handler = r.mws[i](handler)
	}
	return handler
}

// Router2

type info struct {
	static int
	wc     int
	logs   []log
}

type Router2 struct {
	multiplexers map[string]muxInterface
}

func Compile(d *Dune) *Router2 {
	r := &Router2{map[string]muxInterface{}}

	methRoute := make(map[string]*info)

	for _, lg := range *d.logs {
		if _, ok := methRoute[lg.method]; !ok {
			methRoute[lg.method] = &info{}
		}

		inf := methRoute[lg.method]
		inf.logs = append(inf.logs, lg)

		static := strings.IndexFunc(lg.path, func(r rune) bool {
			return r == ':' || r == '*'
		}) == -1

		if static {
			inf.static++
		} else {
			inf.wc++
		}
	}

	for meth, inf := range methRoute {
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

		for _, lg := range inf.logs {
			mux.add(lg.path, lg.handler)
		}

		r.multiplexers[meth] = mux
	}

	return r
}

func (router *Router2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	pc := mux.tree.insert(path, func(w http.ResponseWriter, r *http.Request, p *Params) {
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

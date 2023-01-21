package dune

import (
	"net/http"
	"sort"
	"strings"
	"sync"
)

type Action uint8

const (
	DoNothing Action = iota
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
		logs: &[]log{},
		base: "",
		mws:  nil,
		config: &Config{
			defaultConfig.OnTrailingSlashMatch,
			defaultConfig.OnFixedPathMatch,
			defaultConfig.NotFoundHandler,
			defaultConfig.HandleMethodNotAllowed,
		},
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

type Route struct {
	Method string
	Path   string
}

func (d *Dune) Routes() (routes []Route) {
	routes = make([]Route, 0, len(*d.logs))
	all := len(d.base) == 0

	for _, log := range *d.logs {
		if all || strings.HasPrefix(log.path, d.base) {
			last := len(routes)
			routes = routes[:last+1]
			routes[last] = Route{
				Method: log.method,
				Path:   log.path,
			}
		}
	}

	return
}

// Router

type Router struct {
	muxes       [9]muxInterface         // Stores mux objects for default http methods.
	muxIndices  []int                   // Store indices of non-nil muxes.
	customMuxes map[string]muxInterface // Stores mux objects for custom http methods.
	config      *Config
}

func Compile(d *Dune) *Router {
	type log struct {
		static  bool
		method  string
		path    string
		handler Handler
	}

	type info struct {
		totalStatic int
		logs        []log
	}

	r := &Router{[9]muxInterface{}, nil, nil, d.config}

	methodsInfo := make(map[string]*info)

	// Arrange routes by the http methods. And count static routes.
	for _, lg := range *d.logs {
		if _, ok := methodsInfo[lg.method]; !ok {
			methodsInfo[lg.method] = &info{}
		}

		static := isStatic(lg.path)

		inf := methodsInfo[lg.method]
		inf.logs = append(inf.logs, log{
			static:  static,
			method:  lg.method,
			path:    lg.path,
			handler: lg.handler,
		})

		if static {
			inf.totalStatic++
		}
	}

	// Create a mux variant for each http method based on various parameters and register routes.
	for meth, inf := range methodsInfo {
		var mux muxInterface

		total := len(inf.logs) // Total routes in the method.
		staticPercentage := float64(inf.totalStatic) / float64(total) * 100

		// Determine mux variant.
		if staticPercentage == 100 {
			mux = newStaticMux()
		} else if staticPercentage >= 30 {
			mux = newHybridMux()
		} else {
			mux = newRadixMux()
		}

		// Register routes.
		for _, log := range inf.logs {
			mux.add(log.path, log.static, log.handler)
		}

		// Store mux.
		if idx := methodIndex(meth); idx >= 0 {
			r.muxes[idx] = mux

			// Store indices of active muxes in ascending order.
			r.muxIndices = append(r.muxIndices, idx)
			sort.Slice(r.muxIndices, func(i, j int) bool {
				return r.muxIndices[i] < r.muxIndices[j]
			})
		} else {
			if r.customMuxes == nil {
				r.customMuxes = make(map[string]muxInterface)
			}
			r.customMuxes[meth] = mux
		}
	}

	return r
}

func methodIndex(method string) int {
	switch method {
	case MethodGet:
		return 0
	case MethodPost:
		return 1
	case MethodPut:
		return 2
	case MethodPatch:
		return 3
	case MethodDelete:
		return 4
	case MethodHead:
		return 5
	case MethodOptions:
		return 6
	case MethodTrace:
		return 7
	case MethodConnect:
		return 8
	default:
		return -1
	}
}

func methodString(idx int) string {
	switch idx {
	case 0:
		return MethodGet
	case 1:
		return MethodPost
	case 2:
		return MethodPut
	case 3:
		return MethodPatch
	case 4:
		return MethodDelete
	case 5:
		return MethodHead
	case 6:
		return MethodOptions
	case 7:
		return MethodTrace
	case 8:
		return MethodConnect
	default:
		return ""
	}
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.RawPath
	if len(path) == 0 {
		path = r.URL.Path
	}

	var mux muxInterface
	if idx := methodIndex(r.Method); idx >= 0 {
		mux = rtr.muxes[idx]
	} else {
		mux = rtr.customMuxes[r.Method]
	}

	if mux == nil {
		if rtr.config.HandleMethodNotAllowed {
			rtr.handleMethodNotAllowed(path, r.Method, w)
		}

		rtr.config.NotFoundHandler(w, r, nil)
		return
	}

	handler, ps := mux.find(path)
	if handler != nil {
		if ps == nil {
			ps = emptyParams // Replace with immutable empty params object. Safe for concurrent use.
		}

		handler(w, r, ps)
		return
	}

	// Look with/without trailing slash.
	if rtr.config.OnTrailingSlashMatch != DoNothing {
		var clean string
		if len(path) > 0 && path[len(path)-1] == '/' {
			clean = path[:len(path)-1]
		} else {
			clean = path + "/"
		}

		handler, ps = mux.find(clean)
		if handler != nil {
			switch rtr.config.OnTrailingSlashMatch {
			case DoRedirect:
				r.URL.Path = clean
				http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
				return
			case DoExecute:
				r.URL.Path = clean
				handler(w, r, ps)
				return
			}
		}
	}

	// Clean the path and retry...
	if clean := cleanPath(path); clean != path {
		handler, ps := mux.find(clean)
		if handler == nil {
			if len(clean) > 0 && clean[len(clean)-1] == '/' {
				clean = clean[:len(clean)-1]
			} else {
				clean = clean + "/"
			}

			handler, ps = mux.find(clean)
		}

		if handler != nil {
			switch rtr.config.OnFixedPathMatch {
			case DoRedirect:
				r.URL.Path = clean
				http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
				return
			case DoExecute:
				r.URL.Path = clean
				handler(w, r, ps)
				return
			}
		}
	}

	// Look for allowed methods.
	if rtr.config.HandleMethodNotAllowed {
		rtr.handleMethodNotAllowed(path, r.Method, w)
	}

	rtr.config.NotFoundHandler(w, r, nil)
}

func (rtr *Router) handleMethodNotAllowed(path string, method string, w http.ResponseWriter) {
	allowed := rtr.allowedHeader(path, method)

	if len(allowed) > 0 {
		w.Header().Add("Allow", allowed)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (rtr *Router) allowedHeader(path string, skipMethod string) string {
	var allowed strings.Builder
	skipped := false

	if skipMethodIdx := methodIndex(skipMethod); skipMethodIdx != -1 {
		for _, idx := range rtr.muxIndices {
			if !skipped && idx == skipMethodIdx {
				skipped = true
				continue
			}

			if handler, _ := rtr.muxes[idx].find(path); handler != nil {
				if allowed.Len() != 0 {
					allowed.WriteString(", ")
				}

				allowed.WriteString(methodString(idx))
			}
		}
	}

	for method, mux := range rtr.customMuxes {
		if !skipped && method == skipMethod {
			continue
		}

		if handler, _ := mux.find(path); handler != nil {
			if allowed.Len() != 0 {
				allowed.WriteString(", ")
			}

			allowed.WriteString(method)
		}
	}

	return allowed.String()
}

// Mux

type muxInterface interface {
	add(path string, isStatic bool, handler Handler)
	find(path string) (Handler, *Params)
}

type radixMux struct {
	tree       *node
	paramsPool *sync.Pool
	maxParams  int
}

func newRadixMux() *radixMux {
	return &radixMux{
		tree:       newRootNode(),
		paramsPool: &sync.Pool{},
		maxParams:  0,
	}
}

func (mux *radixMux) add(path string, isStatic bool, handler Handler) {
	if isStatic {
		mux.tree.insert(path, handler)
		return
	}

	// Wrap handler to put Params obj back to the pool after handler execution.
	pc := mux.tree.insert(path, releaseParamsHandler(mux.paramsPool, handler))

	if mux.paramsPool.New == nil || pc > mux.maxParams {
		mux.maxParams = pc
		mux.paramsPool.New = func() any {
			return newParams(pc)
		}
	}
}

// releaseParamsHandler releases the handler's params object into the sync.Pool after execution.
func releaseParamsHandler(pool *sync.Pool, handler Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request, ps *Params) {
		handler(w, r, ps)

		if ps != nil {
			pool.Put(ps)
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

func (mux *staticMux) add(path string, isStatic bool, handler Handler) {
	if !isStatic {
		return
	}

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

func (mux *hybridMux) add(path string, isStatic bool, handler Handler) {
	if isStatic {
		mux.static.add(path, isStatic, handler)
	} else {
		mux.radix.add(path, isStatic, handler)
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

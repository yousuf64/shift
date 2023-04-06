package shift

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type multiplexer interface {
	add(path string, isStatic bool, handler HandlerFunc)
	find(path string) (HandlerFunc, *Params, string)
	findCaseInsensitive(path string, withParams bool) (h HandlerFunc, ps *Params, template string, matchedPath string)
}

// radixMux can store both static and param routes.
// It maps all the routes on a radix tree.
//
// It is recommended to use this multiplexer only when all the routes are param routes.
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

func (mux *radixMux) add(path string, isStatic bool, handler HandlerFunc) {
	// Static routes doesn't need to worry about releasing Params.
	if isStatic {
		mux.tree.insert(path, handler)
		return
	}

	// Wrap request handler by the release params handler. So that Params object is put back to the pool for reuse.
	vc := mux.tree.insert(path, releaseParamsHandler(mux.paramsPool, handler))

	if mux.paramsPool.New == nil || vc > mux.maxParams {
		mux.maxParams = vc
		mux.paramsPool.New = func() interface{} {
			return newParams(vc)
		}
	}
}

// releaseParamsHandler releases the request handler's Params object into the sync.Pool after execution.
// Here the downside is, if panics are not recovered in the chain, this wrapper won't get executed.
// Therefore, Params won't get released to the pool.
func releaseParamsHandler(pool *sync.Pool, handler HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) (err error) {
		err = handler(w, r, route)

		// TODO: Check for emptyParams?
		if route.Params != nil {
			route.Params.reset()
			pool.Put(route.Params)
			route.Params = nil
		}

		return
	}
}

func (mux *radixMux) find(path string) (HandlerFunc, *Params, string) {
	n, ps := mux.tree.search(path, func() *Params {
		ps := mux.paramsPool.Get().(*Params)
		return ps
	})

	if n != nil && n.handler != nil {
		return n.handler, ps, n.template
	}

	return nil, nil, ""
}

func (mux *radixMux) findCaseInsensitive(path string, withParams bool) (HandlerFunc, *Params, string, string) {
	n, ps, matchedPath := mux.tree.caseInsensitiveSearch(path, func() *Params {
		ps := mux.paramsPool.Get().(*Params)
		return ps
	})

	if n != nil && n.handler != nil {
		// When Params object is not required, release it to the pool and return a nil.
		if !withParams && ps != nil {
			ps.reset()
			mux.paramsPool.Put(ps)
			ps = nil
		}

		return n.handler, ps, n.template, matchedPath
	}

	return nil, nil, "", ""
}

// staticMux can store only static routes.
// It maps the routes' request handlers on a builtin map.
// It also maps route length -> route paths in the byLength matrix.
// Which is useful for membership check and case-insensitive search.
//
// Only use this multiplexer only when all the routes are static routes.
type staticMux struct {
	routes   map[string]HandlerFunc
	byLength [][]string // route length -> route paths. Example: 4 (Length) -> /foo, /bar (Paths)
}

func newStaticMux() *staticMux {
	return &staticMux{
		routes:   map[string]HandlerFunc{},
		byLength: make([][]string, 0),
	}
}

func (mux *staticMux) add(path string, isStatic bool, handler HandlerFunc) {
	if !isStatic {
		return
	}

	scanPath(path)

	if len(path) >= len(mux.byLength) {
		// Grow slice.
		mux.byLength = append(mux.byLength, make([][]string, len(path)-len(mux.byLength)+1)...)
	}

	if _, ok := mux.routes[path]; ok {
		panic(fmt.Sprintf("route %s already registered", path))
	}
	mux.routes[path] = handler
	mux.byLength[len(path)] = append(mux.byLength[len(path)], path)
}

func (mux *staticMux) find(path string) (HandlerFunc, *Params, string) {
	if len(path) >= len(mux.byLength) {
		return nil, nil, ""
	}

	if len(mux.byLength[len(path)]) == 0 {
		// Found no paths with the size.
		return nil, nil, ""
	}

	// Lookup the routes map.
	return mux.routes[path], nil, path
}

func (mux *staticMux) findCaseInsensitive(path string, _ bool) (HandlerFunc, *Params, string, string) {
	if len(path) >= len(mux.byLength) {
		return nil, nil, "", ""
	}

	// Retrieve all the paths with the provided path's length.
	if keys := mux.byLength[len(path)]; len(keys) > 0 {
		for _, key := range keys {
			// Find a matching path.
			if lng := longestPrefixCaseInsensitive(key, path); lng == len(path) {
				return mux.routes[key], nil, key, key
			}
		}
	}

	return nil, nil, "", ""
}

// hybridMux can store both static and param routes.
// It maps static routes on a staticMux and param routes on a radixMux.
//
// It is recommended to use this multiplexer when having both static and param routes.
type hybridMux struct {
	static *staticMux
	radix  *radixMux
}

func newHybridMux() *hybridMux {
	return &hybridMux{newStaticMux(), newRadixMux()}
}

func (mux *hybridMux) add(path string, isStatic bool, handler HandlerFunc) {
	if isStatic {
		mux.static.add(path, isStatic, handler)
	} else {
		mux.radix.add(path, isStatic, handler)
	}
}

func (mux *hybridMux) find(path string) (HandlerFunc, *Params, string) {
	if handler, ps, template := mux.static.find(path); handler != nil {
		return handler, ps, template
	}

	return mux.radix.find(path)
}

func (mux *hybridMux) findCaseInsensitive(path string, withParams bool) (HandlerFunc, *Params, string, string) {
	if handler, ps, template, matchedPath := mux.static.findCaseInsensitive(path, withParams); handler != nil {
		return handler, ps, template, matchedPath
	}

	return mux.radix.findCaseInsensitive(path, withParams)
}

func isStatic(path string) bool {
	return strings.IndexFunc(path, func(r rune) bool {
		return r == ':' || r == '*'
	}) == -1
}

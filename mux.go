package ape

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type multiplexer interface {
	add(path string, isStatic bool, handler HandlerFunc)
	find(path string) (HandlerFunc, *Params, string)
	findCaseInsensitive(path string, withParams bool) (h HandlerFunc, ps *Params, matchedPath string)
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

func (mux *radixMux) add(path string, isStatic bool, handler HandlerFunc) {
	if isStatic {
		mux.tree.insert(path, handler)
		return
	}

	// Wrap handler to put Params obj back to the pool after handler execution.
	vc := mux.tree.insert(path, releaseParamsHandler(mux.paramsPool, handler))

	if mux.paramsPool.New == nil || vc > mux.maxParams {
		mux.maxParams = vc
		mux.paramsPool.New = func() interface{} {
			return newParams(vc)
		}
	}
}

// releaseParamsHandler releases the handler's params object into the sync.Pool after execution.
// Here the downside is, if panics are not recovered in the chain, Params won't get released to the pool.
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

func (mux *radixMux) findCaseInsensitive(path string, withParams bool) (HandlerFunc, *Params, string) {
	n, ps, matchedPath := mux.tree.caseInsensitiveSearch(path, func() *Params {
		ps := mux.paramsPool.Get().(*Params)
		return ps
	})

	if n != nil && n.handler != nil {
		// When params obj is not required, just release it to the pool and return a nil.
		if !withParams && ps != nil {
			ps.reset()
			mux.paramsPool.Put(ps)
			ps = nil
		}

		return n.handler, ps, matchedPath
	}

	return nil, nil, ""
}

type staticMux struct {
	routes   map[string]HandlerFunc
	sizePlot [][]string // Size -> Paths. eg:. 4 (Size) -> /foo, /bar (Paths)
}

func newStaticMux() *staticMux {
	return &staticMux{
		routes:   map[string]HandlerFunc{},
		sizePlot: make([][]string, 5),
	}
}

func (mux *staticMux) add(path string, isStatic bool, handler HandlerFunc) {
	if !isStatic {
		return
	}

	scanPath(path)

	if len(path) >= len(mux.sizePlot) {
		// Grow slice.
		mux.sizePlot = append(mux.sizePlot, make([][]string, len(path)-len(mux.sizePlot)+1)...)
	}

	if _, ok := mux.routes[path]; ok {
		panic(fmt.Sprintf("route %s already registered", path))
	}
	mux.routes[path] = handler
	mux.sizePlot[len(path)] = append(mux.sizePlot[len(path)], path)
}

func (mux *staticMux) find(path string) (HandlerFunc, *Params, string) {
	if len(path) >= len(mux.sizePlot) {
		return nil, nil, ""
	}

	if len(mux.sizePlot[len(path)]) == 0 {
		// Found no paths with the size.
		return nil, nil, ""
	}

	// Lookup the routes map.
	return mux.routes[path], nil, path
}

func (mux *staticMux) findCaseInsensitive(path string, _ bool) (HandlerFunc, *Params, string) {
	if len(path) >= len(mux.sizePlot) {
		return nil, nil, ""
	}

	// Retrieve all the paths with the path's length.
	if keys := mux.sizePlot[len(path)]; len(keys) > 0 {
		for _, key := range keys {
			// Find the matching path.
			if lng := longestPrefixCaseInsensitive(key, path); lng == len(path) {
				return mux.routes[key], nil, key
			}
		}
	}

	return nil, nil, ""
}

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

func (mux *hybridMux) findCaseInsensitive(path string, withParams bool) (HandlerFunc, *Params, string) {
	if handler, ps, matchedPath := mux.static.findCaseInsensitive(path, withParams); handler != nil {
		return handler, ps, matchedPath
	}

	return mux.radix.findCaseInsensitive(path, withParams)
}

func isStatic(path string) bool {
	return strings.IndexFunc(path, func(r rune) bool {
		return r == ':' || r == '*'
	}) == -1
}

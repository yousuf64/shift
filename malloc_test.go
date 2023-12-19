//go:build !race

// Skip race detection on memory allocation tests since race detection may increase memory usage
// by 5-10x and execution time by 2-20x which would cause the malloc tests to fail.
// https://go.dev/doc/articles/race_detector

package shift

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

func TestRouter_ServeHTTP_MixedRoutes_Malloc(t *testing.T) {
	r := newTestRouter()

	paths := map[string]string{
		"/users/find":                   http.MethodGet,
		"/users/find/:name":             http.MethodGet,
		"/users/:id/delete":             http.MethodGet,
		"/users/:id/update":             http.MethodGet,
		"/users/groups/:groupId/dump":   http.MethodGet,
		"/users/groups/:groupId/export": http.MethodGet,
		"/users/delete":                 http.MethodGet,
		"/users/all/dump":               http.MethodGet,
		"/users/all/export":             http.MethodGet,
		"/users/any":                    http.MethodGet,

		"/search":                  http.MethodPost,
		"/search/go":               http.MethodPost,
		"/search/go1.html":         http.MethodPost,
		"/search/index.html":       http.MethodPost,
		"/search/:q":               http.MethodPost,
		"/search/:q/go":            http.MethodPost,
		"/search/:q/go1.html":      http.MethodPost,
		"/search/:q/:w/index.html": http.MethodPost,

		"/src/:dest/invalid": http.MethodPut,
		"/src/invalid":       http.MethodPut,
		"/src1/:dest":        http.MethodPut,
		"/src1":              http.MethodPut,

		"/signal-r/:cmd/reflection": http.MethodPatch,
		"/signal-r":                 http.MethodPatch,
		"/signal-r/:cmd":            http.MethodPatch,

		"/query/unknown/pages":         http.MethodHead,
		"/query/:key/:val/:cmd/single": http.MethodHead,
		"/query/:key":                  http.MethodHead,
		"/query/:key/:val/:cmd":        http.MethodHead,
		"/query/:key/:val":             http.MethodHead,
		"/query/unknown":               http.MethodHead,
		"/query/untold":                http.MethodHead,

		"/questions/:index": http.MethodConnect,
		"/questions":        http.MethodConnect,

		"/graphql":     http.MethodDelete,
		"/graph":       http.MethodDelete,
		"/graphql/cmd": http.MethodDelete,

		"/file":        http.MethodDelete,
		"/file/remove": http.MethodDelete,

		"/hero-:name": http.MethodGet,
	}

	for path, meth := range paths {
		r.Map([]string{meth}, path, fakeHandler())
	}

	tt := []srvTestItem{
		{method: http.MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: http.MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: http.MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		{method: http.MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},

		{method: http.MethodPost, path: "/search", valid: true, pathTemplate: "/search"},
		{method: http.MethodPost, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodPost, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodPost, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		{method: http.MethodPost, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: http.MethodPost, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodPost, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: http.MethodPost, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodPost, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: http.MethodPut, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodPut, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		{method: http.MethodPut, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodPut, path: "/src1", valid: true, pathTemplate: "/src1"},

		{method: http.MethodPatch, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodPatch, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodPatch, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodPatch, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: http.MethodHead, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodHead, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: http.MethodHead, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodHead, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodHead, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodHead, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodHead, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},

		{method: http.MethodConnect, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodConnect, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: http.MethodDelete, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodDelete, path: "/graph", valid: true, pathTemplate: "/graph"},
		{method: http.MethodDelete, path: "/graphql/cmd", valid: true, pathTemplate: "/graphql/cmd", params: nil},

		{method: http.MethodDelete, path: "/file", valid: true, pathTemplate: "/file", params: nil},
		{method: http.MethodDelete, path: "/file/remove", valid: true, pathTemplate: "/file/remove", params: nil},

		{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
	}

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	srv := r.Serve()

	allocations := testing.AllocsPerRun(1000, func() {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	})

	assert(t, allocations == 0, fmt.Sprintf("expected zero allocations, got %g allocations", allocations))
}

func TestRouter_20Params_Malloc(t *testing.T) {
	r := New()

	var template strings.Builder
	var path strings.Builder
	var paramKeys []string

	for i := 1; i <= 20; i++ {
		template.WriteString(fmt.Sprintf("/:%d", i))
		path.WriteString("/foo")
		paramKeys = append(paramKeys, fmt.Sprintf("%d", i))
	}

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		for _, key := range paramKeys {
			if v := route.Params.Get(key); v != "foo" {
				panic(fmt.Sprintf("param value > expected: foo, got: %s", v))
			}
		}
		return nil
	}

	r.GET(template.String(), f)

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)

	allocs := testing.AllocsPerRun(1000, func() {
		srv.ServeHTTP(rw, req)
		if rw.Code != http.StatusOK {
			panic(fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		}
	})

	assert(t, allocs == 0, fmt.Sprintf("allocs > expected: 0, got: %g", allocs))
}

func TestRouter_200Params_Malloc(t *testing.T) {
	r := New()

	var template strings.Builder
	var path strings.Builder
	var paramKeys []string

	for i := 1; i <= 200; i++ {
		template.WriteString(fmt.Sprintf("/:%d", i))
		path.WriteString("/foo")
		paramKeys = append(paramKeys, fmt.Sprintf("%d", i))
	}

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		for _, key := range paramKeys {
			if v := route.Params.Get(key); v != "foo" {
				panic(fmt.Sprintf("param value > expected: foo, got: %s", v))
			}
		}
		return nil
	}

	r.GET(template.String(), f)

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)

	allocs := testing.AllocsPerRun(1000, func() {
		srv.ServeHTTP(rw, req)
		if rw.Code != http.StatusOK {
			panic(fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		}
	})

	assert(t, allocs == 0, fmt.Sprintf("allocs > expected: 0, got: %g", allocs))
}

func TestRouter_2000Params_Malloc(t *testing.T) {
	r := New()

	var template strings.Builder
	var path strings.Builder
	var paramKeys []string

	for i := 1; i <= 2000; i++ {
		template.WriteString(fmt.Sprintf("/:%d", i))
		path.WriteString("/foo")
		paramKeys = append(paramKeys, fmt.Sprintf("%d", i))
	}

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		for _, key := range paramKeys {
			if v := route.Params.Get(key); v != "foo" {
				panic(fmt.Sprintf("param value > expected: foo, got: %s", v))
			}
		}
		return nil
	}

	r.GET(template.String(), f)

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)

	allocs := testing.AllocsPerRun(1000, func() {
		srv.ServeHTTP(rw, req)
		if rw.Code != http.StatusOK {
			panic(fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		}
	})

	assert(t, allocs == 0, fmt.Sprintf("allocs > expected: 0, got: %g", allocs))
}

func TestCleanPath_Malloc(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping malloc count in short mode")
	}

	for _, test := range cleanTests {
		test := test
		allocs := testing.AllocsPerRun(100, func() {
			cleanPath(test.result)
		})
		if allocs > 0 {
			t.Errorf("CleanPath(%q): %v allocs, want zero", test.result, allocs)
		}
	}
}

func TestRouteContextMiddleware_Malloc(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.18") {
		return
	}

	r := New()
	r.Use(RouteContext())
	r.GET("/movies/genres/:name", HTTPHandlerFunc(fakeHttpHandler))
	srv := r.Serve()

	rr := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/movies/genres/western", nil)

	allocs := testing.AllocsPerRun(1000, func() {
		srv.ServeHTTP(rr, req)
	})

	assert(t, allocs == 1, fmt.Sprintf("allocations > expected: %d, got: %g", 1, allocs))
	t.Log(allocs)
}

func TestContext_FromContext_Malloc(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "/movies/111/segments/222/frames/333", nil)
	ctx := req.Context()

	ip := newInternalParams(3)
	ip.setKeys(&[]string{"id", "segmentId", "frameId"})
	ip.appendValue("111")
	ip.appendValue("222")
	ip.appendValue("333")

	req = req.WithContext(WithRoute(ctx, Route{
		Params: newParams(ip),
		Path:   "/movies/:id/segments/:segmentId/frames/:frameId",
	}))

	allocs := testing.AllocsPerRun(1000, func() {
		FromContext(req.Context())
	})

	assert(t, allocs == 0, fmt.Sprintf("allocations > expected: %d, got: %g", 0, allocs))
}

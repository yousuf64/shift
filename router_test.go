package ape

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
)

type routerTestItem1 struct {
	method       string
	path         string
	valid        bool
	pathTemplate string
	params       map[string]string
}

type routerTestItem2 struct {
	method       string
	path         string
	valid        bool
	pathTemplate string
	params       map[string]string
	paramsCount  int
}

type routerScenario = []routerTestItem1

type routerTestTable2 = []routerTestItem2

type mockRW struct {
	http.ResponseWriter

	code int
}

func (m *mockRW) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockRW) WriteHeader(statusCode int) {
	m.code = statusCode
}

func (m *mockRW) Header() http.Header {
	return map[string][]string{}
}

func newTestDune() *Router {
	return New()
}

func TestRouter_ServeHTTP_StaticRoutes(t *testing.T) {
	r := newTestDune()

	paths := map[string]string{
		"/users/find":          http.MethodGet,
		"/users/delete":        http.MethodGet,
		"/users/all/dump":      http.MethodGet,
		"/users/all/export":    http.MethodGet,
		"/users/any":           http.MethodGet,
		"/search":              http.MethodGet,
		"/search/go":           http.MethodGet,
		"/search/go1.html":     http.MethodGet,
		"/search/index.html":   http.MethodGet,
		"/src/invalid":         http.MethodGet,
		"/src1":                http.MethodGet,
		"/signal-r":            http.MethodGet,
		"/query/unknown":       http.MethodGet,
		"/query/unknown/pages": http.MethodGet,
		"/query/untold":        http.MethodGet,
		"/questions":           http.MethodGet,
		"/graphql":             http.MethodGet,
		"/graph":               http.MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler(path))
	}

	tt := routerScenario{
		{method: http.MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		{method: http.MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		{method: http.MethodGet, path: "/users/911", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		{method: http.MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		{method: http.MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/search/contact.html", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		{method: http.MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		{method: http.MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodGet, path: "/signal-r/connect", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		{method: http.MethodGet, path: "/query", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},
		{method: http.MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		{method: http.MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
	}

	testRouter_ServeHTTP(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_ParamRoutes(t *testing.T) {
	r := newTestDune()

	paths := map[string]string{
		"/users/find/:name":             http.MethodGet,
		"/users/:id/delete":             http.MethodDelete,
		"/users/groups/:groupId/dump":   http.MethodPut,
		"/users/groups/:groupId/export": http.MethodGet,
		"/users/:id/update":             http.MethodPut,
		"/search/:q":                    http.MethodGet,
		"/search/:q/go":                 http.MethodGet,
		"/search/:q/go1.html":           http.MethodTrace,
		"/search/:q/:w/index.html":      http.MethodGet,
		"/src/:dest/invalid":            http.MethodGet,
		"/src1/:dest":                   http.MethodGet,
		"/signal-r/:cmd":                http.MethodGet,
		"/signal-r/:cmd/reflection":     http.MethodGet,
		"/query/:key":                   http.MethodGet,
		"/query/:key/:val":              http.MethodGet,
		"/query/:key/:val/:cmd":         http.MethodGet,
		"/query/:key/:val/:cmd/single":  http.MethodGet,
		"/questions/:index":             http.MethodGet,
		"/graphql/:cmd":                 http.MethodPut,
		"/:file":                        http.MethodGet,
		"/:file/remove":                 http.MethodGet,
		"/hero-:name":                   http.MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler(path))
	}

	tt := routerScenario{
		{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: http.MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodDelete, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodPut, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: http.MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodPut, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: http.MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodTrace, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},
		{method: http.MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: http.MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodPut, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		{method: http.MethodPut, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: http.MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		{method: http.MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		{method: http.MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	testRouter_ServeHTTP(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_DifferentParamNames(t *testing.T) {
	r := newTestDune()

	f := func(kvs map[string]string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			for k, v := range kvs {
				val := route.Params.Get(k)
				assert(t, v == val, fmt.Sprintf("%s for param %s > expected: %s, got: %s", r.URL.String(), k, v, val))
			}
			return nil
		}
	}

	r1, _ := http.NewRequest(http.MethodGet, "/foo/911", nil)
	r.GET("/foo/:id", f(map[string]string{"id": "911"}))

	r2, _ := http.NewRequest(http.MethodGet, "/foo/bar/abc", nil)
	r.GET("/foo/:name/abc", f(map[string]string{"name": "bar"}))

	r3, _ := http.NewRequest(http.MethodGet, "/xyzooo", nil)
	r.GET("/xyz:param", f(map[string]string{"param": "ooo"}))

	r4, _ := http.NewRequest(http.MethodGet, "/xyzgo/aaa", nil)
	r.GET("/xyz:lang/aaa", f(map[string]string{"lang": "go"}))

	r5, _ := http.NewRequest(http.MethodGet, "/www/dune/jpeg", nil)
	r.GET("/www/:filename/:extension", f(map[string]string{"filename": "dune", "extension": "jpeg"}))

	r6, _ := http.NewRequest(http.MethodGet, "/www/meme/gif/upload", nil)
	r.GET("/www/:file/:ext/upload", f(map[string]string{"file": "meme", "ext": "gif"}))

	svr := r.Serve()

	requests := [...]*http.Request{r1, r2, r3, r4, r5, r6}
	rw := httptest.NewRecorder()

	for _, req := range requests {
		svr.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK, fmt.Sprintf("%p http status > expected: %d, got: %d", req.URL, http.StatusOK, rw.Code))
	}
}

func TestRouter_ServeHTTP_WildcardRoutes(t *testing.T) {
	r := newTestDune()

	paths := map[string]string{
		"/messages/*action":     http.MethodGet,
		"/users/posts/*command": http.MethodGet,
		"/images/*filepath":     http.MethodGet,
		"/hero-*dir":            http.MethodGet,
		"/netflix*abc":          http.MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler(path))
	}

	tt := routerScenario{
		{method: http.MethodGet, path: "/messages/publish", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish"}},
		{method: http.MethodGet, path: "/messages/publish/OrderPlaced", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish/OrderPlaced"}},
		{method: http.MethodGet, path: "/messages/", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": ""}},
		{method: http.MethodGet, path: "/messages", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/posts/", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": ""}},
		{method: http.MethodGet, path: "/users/posts", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/posts/push", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push"}},
		{method: http.MethodGet, path: "/users/posts/push/911", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push/911"}},
		{method: http.MethodGet, path: "/images/gopher.png", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "gopher.png"}},
		{method: http.MethodGet, path: "/images/", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": ""}},
		{method: http.MethodGet, path: "/images", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/images/svg/up-icon", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "svg/up-icon"}},
		{method: http.MethodGet, path: "/hero-dc/batman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/batman.json"}},
		{method: http.MethodGet, path: "/hero-dc/superman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/superman.json"}},
		{method: http.MethodGet, path: "/hero-marvel/loki.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "marvel/loki.json"}},
		{method: http.MethodGet, path: "/hero-", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": ""}},
		{method: http.MethodGet, path: "/hero", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/netflix", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": ""}},
		{method: http.MethodGet, path: "/netflix++", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": "++"}},
		{method: http.MethodGet, path: "/netflix/drama/better-call-saul", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": "/drama/better-call-saul"}},
	}

	testRouter_ServeHTTP(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_MixedRoutes(t *testing.T) {
	d := newTestDune()

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

		"/search":                  http.MethodGet,
		"/search/go":               http.MethodGet,
		"/search/go1.html":         http.MethodGet,
		"/search/index.html":       http.MethodGet,
		"/search/:q":               http.MethodGet,
		"/search/:q/go":            http.MethodGet,
		"/search/:q/go1.html":      http.MethodGet,
		"/search/:q/:w/index.html": http.MethodGet,

		"/src/:dest/invalid": http.MethodGet,
		"/src/invalid":       http.MethodGet,
		"/src1/:dest":        http.MethodGet,
		"/src1":              http.MethodGet,

		"/signal-r/:cmd/reflection": http.MethodGet,
		"/signal-r":                 http.MethodGet,
		"/signal-r/:cmd":            http.MethodGet,

		"/query/unknown/pages":         http.MethodGet,
		"/query/:key/:val/:cmd/single": http.MethodGet,
		"/query/:key":                  http.MethodGet,
		"/query/:key/:val/:cmd":        http.MethodGet,
		"/query/:key/:val":             http.MethodGet,
		"/query/unknown":               http.MethodGet,
		"/query/untold":                http.MethodGet,

		"/questions/:index": http.MethodGet,
		"/questions":        http.MethodGet,

		"/graphql":      http.MethodGet,
		"/graph":        http.MethodGet,
		"/graphql/:cmd": http.MethodGet,

		"/:file":        http.MethodGet,
		"/:file/remove": http.MethodGet,

		"/hero-:name": http.MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		d.Map([]string{meth}, path, rec.Handler(path))
	}

	tt := routerScenario{
		{method: http.MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: http.MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: http.MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		{method: http.MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		{method: http.MethodGet, path: "/users/911", valid: false, pathTemplate: ""},

		{method: http.MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		{method: http.MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		{method: http.MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: http.MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: http.MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: http.MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		{method: http.MethodGet, path: "/src", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "src"}},
		{method: http.MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		{method: http.MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: http.MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: http.MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		{method: http.MethodGet, path: "/query", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "query"}},

		{method: http.MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: http.MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		{method: http.MethodGet, path: "/graphq", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "graphq"}},
		{method: http.MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		{method: http.MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: http.MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		{method: http.MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		{method: http.MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	testRouter_ServeHTTP(t, d.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_FallbackToParamRoute(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/search/got":           http.MethodGet,
			"/search/:q":            http.MethodGet,
			"/search/:q/go":         http.MethodGet,
			"/search/:q/go/*action": http.MethodGet,
			"/search/:q/*action":    http.MethodGet,
			"/search/*action":       http.MethodGet, // Should never be matched, since it's overridden by a param segment (/search/:q/*action) whose next segment is a wildcard segment.
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/search/gotten", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			{method: http.MethodGet, path: "/search/got", valid: true, pathTemplate: "/search/got", paramsCount: 0, params: nil},
			{method: http.MethodGet, path: "/search/gopher", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gopher"}},
			{method: http.MethodGet, path: "/search/gopher/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gopher"}},
			{method: http.MethodGet, path: "/search/gok", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gok"}},
			{method: http.MethodGet, path: "/search/gok/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gok"}},
			{method: http.MethodGet, path: "/search/gotten/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			{method: http.MethodGet, path: "/search/gotten/goner", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "gotten", "action": "goner"}},
			{method: http.MethodGet, path: "/search/gotham", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotham"}},
			{method: http.MethodGet, path: "/search/got/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "got"}},
			{method: http.MethodGet, path: "/search/got/gone", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "gone"}},
			{method: http.MethodGet, path: "/search/gotham/joker", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "gotham", "action": "joker"}},
			{method: http.MethodGet, path: "/search/got/go/pro", valid: true, pathTemplate: "/search/:q/go/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "pro"}},
			{method: http.MethodGet, path: "/search/got/go/", valid: true, pathTemplate: "/search/:q/go/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": ""}},
			{method: http.MethodGet, path: "/search/got/apple", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "apple"}},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/search/go/go/goose":   http.MethodGet,
			"/search/:q":            http.MethodGet,
			"/search/:q/go/goos:x":  http.MethodGet,
			"/search/:q/g:w/goos:x": http.MethodGet,
			"/search/:q/go":         http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/search/gotten", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			{method: http.MethodGet, path: "/search/gox", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gox"}},
			{method: http.MethodGet, path: "/search/go/go/goose", valid: true, pathTemplate: "/search/go/go/goose", paramsCount: 0, params: nil},
			{method: http.MethodGet, path: "/search/go/go/goosf", valid: true, pathTemplate: "/search/:q/go/goos:x", paramsCount: 2, params: map[string]string{"q": "go", "x": "f"}},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_FallbackToWildcardRoute(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/search/:q/stop": http.MethodGet,
			"/search/*action": http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/search/cherry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry"}},
			{method: http.MethodGet, path: "/search/cherry/", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/"}},
			{method: http.MethodGet, path: "/search/cherry/berry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/berry"}},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/foo/apple/mango/:fruit": http.MethodGet,
			"/foo/*tag":               http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/foo/apple/orange", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/orange"}},
			{method: http.MethodGet, path: "/foo/apple/mango/pineapple/another-fruit", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/mango/pineapple/another-fruit"}},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)

	})

	t.Run("scenario 3", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/foo/apple/mango/hanna":  http.MethodGet,
			"/foo/apple/mango/:fruit": http.MethodGet,
			"/foo/*tag":               http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/foo/apple/mango/hanna-banana", valid: true, pathTemplate: "/foo/apple/mango/:fruit", paramsCount: 1, params: map[string]string{"fruit": "hanna-banana"}},
			{method: http.MethodGet, path: "/foo/apple/mango/hanna-banana/watermelon", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/mango/hanna-banana/watermelon"}},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_RecordParamsOnlyForMatchedPath(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/search":             http.MethodGet,
			"/search/:q/stop":     http.MethodGet,
			"/search/*action":     http.MethodGet,
			"/geo/:lat/:lng/path": http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/search/cherry/", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/"}},
			{method: http.MethodGet, path: "/search/cherry/berry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/berry"}},
			{method: http.MethodGet, path: "/geo/135/280/path/optimize", valid: false, pathTemplate: "", paramsCount: 0, params: nil},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/search/go":        http.MethodGet,
			"/search/:var/tail": http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: http.MethodGet, path: "/search/gopher", valid: false, pathTemplate: "", paramsCount: 0, params: nil},
		}

		testRouter_ServeHTTP_2(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_Priority(t *testing.T) {
	t.Run("static > wildcard", func(t *testing.T) {
		d := newTestDune()

		paths := map[string]string{
			"/better-call-saul_":        http.MethodGet,
			"/better-call-saul_*season": http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerScenario{
			{method: http.MethodGet, path: "/better-call-saul_", valid: true, pathTemplate: "/better-call-saul_", params: nil},
			{method: http.MethodGet, path: "/better-call-saul_6", valid: true, pathTemplate: "/better-call-saul_*season", params: map[string]string{"season": "6"}},
		}

		testRouter_ServeHTTP(t, d.Serve(), rec, tt)
	})

	t.Run("param > wildcard", func(t *testing.T) {
		r := newTestDune()

		paths := map[string]string{
			"/dark_:season": http.MethodGet,
			"/dark_*wc":     http.MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler(path))
		}

		tt := routerScenario{
			{method: http.MethodGet, path: "/dark_3", valid: true, pathTemplate: "/dark_:season", params: map[string]string{"season": "3"}},
			{method: http.MethodGet, path: "/dark_", valid: true, pathTemplate: "/dark_*wc", params: map[string]string{"wc": ""}},
		}

		testRouter_ServeHTTP(t, r.Serve(), rec, tt)
	})

}

func TestRouter_ServeHTTP_CaseInsensitive(t *testing.T) {
	r := New()
	r.UseSanitizeURLMatch(WithExecute())

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

		"/questions/:index": http.MethodOptions,
		"/questions":        http.MethodOptions,

		"/graphql":     http.MethodDelete,
		"/graph":       http.MethodDelete,
		"/graphql/cmd": http.MethodDelete,

		"/file":        http.MethodDelete,
		"/file/remove": http.MethodDelete,

		"/hero-:name": http.MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler(path))
	}

	tt := routerScenario{
		{method: http.MethodGet, path: "/users/finD", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/finD/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: http.MethodGet, path: "/users/john/deletE", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodGet, path: "/users/911/updatE", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/users/groupS/120/dumP", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groupS/230/exporT", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: http.MethodGet, path: "/users/deletE", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodGet, path: "/users/alL/dumP", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodGet, path: "/users/alL/exporT", valid: true, pathTemplate: "/users/all/export"},
		{method: http.MethodGet, path: "/users/AnY", valid: true, pathTemplate: "/users/any"},

		{method: http.MethodPost, path: "/seArcH", valid: true, pathTemplate: "/search"},
		{method: http.MethodPost, path: "/sEarCh/gO", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodPost, path: "/SeArcH/Go1.hTMl", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodPost, path: "/sEaRch/inDEx.hTMl", valid: true, pathTemplate: "/search/index.html"},
		{method: http.MethodPost, path: "/SEARCH/contact.html", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "contact.html"}},
		{method: http.MethodPost, path: "/SeArCh/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodPost, path: "/sEArCH/gophers/Go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: http.MethodPost, path: "/sEArCH/nature/go1.HTML", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodPost, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: http.MethodPut, path: "/Src/paris/InValiD", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodPut, path: "/SrC/InvaliD", valid: true, pathTemplate: "/src/invalid"},
		{method: http.MethodPut, path: "/SrC1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodPut, path: "/SrC1", valid: true, pathTemplate: "/src1"},

		{method: http.MethodPatch, path: "/Signal-R/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodPatch, path: "/sIgNaL-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodPatch, path: "/SIGNAL-R/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodPatch, path: "/sIGNal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: http.MethodHead, path: "/quERy/unKNown/paGEs", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodHead, path: "/QUery/10/amazing/reset/SiNglE", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: http.MethodHead, path: "/QueRy/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodHead, path: "/qUERy/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodHead, path: "/QueRy/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodHead, path: "/qUeRy/uNkNoWn", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodHead, path: "/QuerY/UntOld", valid: true, pathTemplate: "/query/untold"},

		{method: http.MethodOptions, path: "/qUestions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodOptions, path: "/quEsTioNs", valid: true, pathTemplate: "/questions"},

		{method: http.MethodDelete, path: "/GRAPHQL", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodDelete, path: "/gRapH", valid: true, pathTemplate: "/graph"},
		{method: http.MethodDelete, path: "/grAphQl/cMd", valid: true, pathTemplate: "/graphql/cmd", params: nil},

		{method: http.MethodDelete, path: "/File", valid: true, pathTemplate: "/file", params: nil},
		{method: http.MethodDelete, path: "/fIle/rEmOve", valid: true, pathTemplate: "/file/remove", params: nil},

		{method: http.MethodGet, path: "/heRO-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: http.MethodGet, path: "/HEro-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
	}

	testRouter_ServeHTTP(t, r.Serve(), rec, tt)
}

func TestStaticMux_CaseInsensitiveSearch(t *testing.T) {
	r := New()
	r.UseSanitizeURLMatch(WithExecute())

	f := func(path string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			u := r.URL.String()
			assert(t, path == u, fmt.Sprintf("request path expected: %s, got: %s", path, u))
			return nil
		}
	}

	r.GET("/foo", f("/foo"))
	r.GET("/abc/xyz", f("/abc/xyz"))
	r.GET("/go/go/go", f("/go/go/go"))
	r.GET("/bar/ccc/ddd/zzz", f("/bar/ccc/ddd/zzz"))

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/FoO", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/aBC/xYz", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/gO/GO/go", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/BaR/CcC/ddD/zzZ", nil)

	requests := [...]*http.Request{r1, r2, r3, r4}
	rw := httptest.NewRecorder()

	for _, request := range requests {
		srv.ServeHTTP(rw, request)
		assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status expected: %d, got: %d", http.StatusOK, rw.Code))
	}
}

func testRouter_ServeHTTP(t *testing.T, r *Server, rec *recorder, table routerScenario) {
	for _, tx := range table {
		rw := &mockRW{}
		req, _ := http.NewRequest(tx.method, tx.path, nil)

		r.ServeHTTP(rw, req)

		ok := rw.code == 0
		notFound := rw.code == 404

		assertOn(t, tx.valid, ok, fmt.Sprintf("%s > expected a handler, but didn't find a handler", tx.path))
		assertOn(t, !tx.valid, notFound, fmt.Sprintf("%s > didn't expect a handler, but found a handler with template '%s'", tx.path, rec.path))
		assertOn(t, tx.valid && ok, rec.path == tx.pathTemplate, fmt.Sprintf("%s > path template expected: %s, got: %s", tx.path, tx.pathTemplate, rec.path))

		if tx.valid && ok {
			for k, v := range tx.params {
				actual := rec.params.Get(k)
				assert(t, actual == v, fmt.Sprintf("%s > param '%s' > expected value: '%s', got '%s'", tx.path, k, v, actual))
			}
		}
	}
}

func testRouter_ServeHTTP_2(t *testing.T, r *Server, rec *recorder, table routerTestTable2) {
	for _, tx := range table {
		rw := &mockRW{}
		req, _ := http.NewRequest(tx.method, tx.path, nil)

		r.ServeHTTP(rw, req)

		ok := rw.code == 0
		notFound := rw.code == 404

		assertOn(t, tx.valid, ok, fmt.Sprintf("%s > expected a handler, but didn't find a handler", tx.path))
		assertOn(t, !tx.valid, notFound, fmt.Sprintf("%s > didn't expect a handler, but found a handler with template '%s'", tx.path, rec.path))
		assertOn(t, tx.valid && ok, rec.path == tx.pathTemplate, fmt.Sprintf("%s > path template expected: %s, got: %s", tx.path, tx.pathTemplate, rec.path))

		gotParamsCount := 0
		if rec.params != nil {
			gotParamsCount = len(rec.params.values)
		}
		assert(t, tx.paramsCount == gotParamsCount, fmt.Sprintf("%s > params count expected: %d, got: %d", tx.path, tx.paramsCount, gotParamsCount))

		if len(tx.params) > 0 {
			for k, v := range tx.params {
				actual := rec.params.Get(k)
				assert(t, actual == v, fmt.Sprintf("%s > param '%s' > expected value: '%s', got '%s'", tx.path, k, v, actual))
			}
		}

		// Reset recorder...
		rec.path = ""
		rec.params = nil
	}
}

func BenchmarkRouter_ServeHTTP_StaticRoutes(b *testing.B) {
	r := newTestDune()

	routes := []string{
		"/users/find",
		"/users/delete",
		"/users/all/dump",
		"/users/all/export",
		"/users/any",
		"/search",
		"/search/go",
		"/search/go1.html",
		"/search/index.html",
		"/src/invalid",
		"/src1",
		"/signal-r",
		"/query/unknown",
		"/query/unknown/pages",
		"/query/untold",
		"/questions",
		"/graphql",
		"/graph/var",
		//"/graph/:var",
	}

	testers := []string{
		"/users/find",
		"/users/delete",
		"/users/all/dump",
		"/users/all/export",
		"/users/any",
		"/search",
		"/search/go",
		"/search/go1.html",
		"/search/index.html",
		"/src/invalid",
		"/src1",
		"/signal-r",
		"/query/unknown",
		"/query/unknown/pages",
		"/query/untold",
		"/questions",
		"/graphql",
		"/graph/var",
		//"/graph/2000",
	}

	for _, route := range routes {
		r.GET(route, fakeHandler())
	}

	requests := make([]*http.Request, len(testers))

	for i, path := range testers {
		requests[i], _ = http.NewRequest("GET", path, nil)
	}

	srv := r.Serve()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_MixedRoutes_S(b *testing.B) {
	r := newTestDune()

	routes := []string{
		"/posts",
		"/posts/ants",
		"/posts/antonio/cesaro",
		"/skus",
		"/skus/:id",
		"/skus/:id/categories",
		"/skus/:id/categories/:cid",
		"/skus/:id/categories/all",
		"/skus/:id/categories/one",
	}

	testers := []string{
		"/posts",
		"/posts/ants",
		"/posts/antonio/cesaro",
		"/skus",
		"/skus/123",
		"/skus/123/categories",
		"/skus/123/categories/899",
		"/skus/123/categories/all",
		"/skus/123/categories/one",
	}

	for _, route := range routes {
		r.GET(route, fakeHandler())
	}

	requests := make([]*http.Request, len(testers))

	for i, path := range testers {
		requests[i], _ = http.NewRequest("GET", path, nil)
	}

	srv := r.Serve()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_MixedRoutes_X_1(b *testing.B) {
	r := newTestDune()

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

		"/search":                  http.MethodGet,
		"/search/go":               http.MethodGet,
		"/search/go1.html":         http.MethodGet,
		"/search/index.html":       http.MethodGet,
		"/search/:q":               http.MethodGet,
		"/search/:q/go":            http.MethodGet,
		"/search/:q/go1.html":      http.MethodGet,
		"/search/:q/:w/index.html": http.MethodGet,

		"/src/:dest/invalid": http.MethodGet,
		"/src/invalid":       http.MethodGet,
		"/src1/:dest":        http.MethodGet,
		"/src1":              http.MethodGet,

		"/signal-r/:cmd/reflection": http.MethodGet,
		"/signal-r":                 http.MethodGet,
		"/signal-r/:cmd":            http.MethodGet,

		"/query/unknown/pages":         http.MethodGet,
		"/query/:key/:val/:cmd/single": http.MethodGet,
		"/query/:key":                  http.MethodGet,
		"/query/:key/:val/:cmd":        http.MethodGet,
		"/query/:key/:val":             http.MethodGet,
		"/query/unknown":               http.MethodGet,
		"/query/untold":                http.MethodGet,

		"/questions/:index": http.MethodGet,
		"/questions":        http.MethodGet,

		"/graphql":      http.MethodGet,
		"/graph":        http.MethodGet,
		"/graphql/:cmd": http.MethodGet,

		"/:file":        http.MethodGet,
		"/:file/remove": http.MethodGet,

		//"/hero-:name": http.MethodGet,
	}

	for path, meth := range paths {
		r.Map([]string{meth}, path, fakeHandler())
	}

	tt := routerScenario{
		//{method: http.MethodGet, path: "/src", valid: false, pathTemplate: ""},

		{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},

		{method: http.MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		//{method: http.MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		//{method: http.MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		//{method: http.MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		//{method: http.MethodGet, path: "/users/911", valid: false, pathTemplate: ""},

		{method: http.MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		{method: http.MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		//{method: http.MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: http.MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		//{method: http.MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: http.MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		//{method: http.MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		//{method: http.MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: http.MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		//{method: http.MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		//{method: http.MethodGet, path: "/query", valid: false, pathTemplate: ""},

		{method: http.MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: http.MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		//{method: http.MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		//{method: http.MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: http.MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		//{method: http.MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		//{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		//{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		//{method: http.MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	srv := r.Serve()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_MixedRoutes_X_2(b *testing.B) {
	r := newTestDune()

	paths := map[string]string{
		"/users/find":                   http.MethodPost,
		"/users/find/:name":             http.MethodGet,
		"/users/:id/delete":             http.MethodGet,
		"/users/:id/update":             http.MethodGet,
		"/users/groups/:groupId/dump":   http.MethodGet,
		"/users/groups/:groupId/export": http.MethodGet,
		"/users/delete":                 http.MethodPost,
		"/users/all/dump":               http.MethodPost,
		"/users/all/export":             http.MethodPost,
		"/users/any":                    http.MethodPost,

		"/search":                  http.MethodPost,
		"/search/go":               http.MethodPost,
		"/search/go1.html":         http.MethodPost,
		"/search/index.html":       http.MethodPost,
		"/search/:q":               http.MethodGet,
		"/search/:q/go":            http.MethodGet,
		"/search/:q/go1.html":      http.MethodGet,
		"/search/:q/:w/index.html": http.MethodGet,

		"/src/:dest/invalid": http.MethodGet,
		"/src/invalid":       http.MethodPost,
		"/src1/:dest":        http.MethodGet,
		"/src1":              http.MethodPost,

		"/signal-r/:cmd/reflection": http.MethodGet,
		"/signal-r":                 http.MethodPost,
		"/signal-r/:cmd":            http.MethodGet,

		"/query/unknown/pages":         http.MethodPost,
		"/query/:key/:val/:cmd/single": http.MethodGet,
		"/query/:key":                  http.MethodGet,
		"/query/:key/:val/:cmd":        http.MethodGet,
		"/query/:key/:val":             http.MethodGet,
		"/query/unknown":               http.MethodPost,
		"/query/untold":                http.MethodPost,

		"/questions/:index": http.MethodGet,
		"/questions":        http.MethodPost,

		"/graphql":      http.MethodPost,
		"/graph":        http.MethodPost,
		"/graphql/:cmd": http.MethodGet,

		"/:file":        http.MethodGet,
		"/:file/remove": http.MethodGet,

		//"/hero-:name": http.MethodGet,
	}

	for path, meth := range paths {
		r.Map([]string{meth}, path, fakeHandler())
	}

	tt := routerScenario{
		//{method: http.MethodGet, path: "/src", valid: false, pathTemplate: ""},

		{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},

		{method: http.MethodPost, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		//{method: http.MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		//{method: http.MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodPost, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodPost, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodPost, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		//{method: http.MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: http.MethodPost, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		//{method: http.MethodGet, path: "/users/911", valid: false, pathTemplate: ""},

		{method: http.MethodPost, path: "/search", valid: true, pathTemplate: "/search"},
		{method: http.MethodPost, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodPost, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodPost, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		//{method: MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: http.MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		//{method: MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: http.MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodPost, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		//{method: http.MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodPost, path: "/src1", valid: true, pathTemplate: "/src1"},
		//{method: http.MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodPost, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: http.MethodPost, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		//{method: http.MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: http.MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodPost, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodPost, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		//{method: http.MethodGet, path: "/query", valid: false, pathTemplate: ""},

		{method: http.MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodPost, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: http.MethodPost, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodPost, path: "/graph", valid: true, pathTemplate: "/graph"},
		//{method: http.MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
		{method: http.MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		//{method: http.MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: http.MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: http.MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		//{method: http.MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		//{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		//{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		//{method: http.MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	srv := r.Serve()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_MixedRoutes_X_3(b *testing.B) {
	r := newTestDune()

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

	tt := routerScenario{
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

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_CaseInsensitive(b *testing.B) {
	r := New()
	r.UseSanitizeURLMatch(WithRedirect())

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

		"/questions/:index": http.MethodOptions,
		"/questions":        http.MethodOptions,

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

	tt := routerScenario{
		{method: http.MethodGet, path: "/users/finD", valid: true, pathTemplate: "/users/find"},
		{method: http.MethodGet, path: "/users/finD/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: http.MethodGet, path: "/users/john/deletE", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: http.MethodGet, path: "/users/911/updatE", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: http.MethodGet, path: "/users/groupS/120/dumP", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: http.MethodGet, path: "/users/groupS/230/exporT", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: http.MethodGet, path: "/users/deletE", valid: true, pathTemplate: "/users/delete"},
		{method: http.MethodGet, path: "/users/alL/dumP", valid: true, pathTemplate: "/users/all/dump"},
		{method: http.MethodGet, path: "/users/alL/exporT", valid: true, pathTemplate: "/users/all/export"},
		{method: http.MethodGet, path: "/users/AnY", valid: true, pathTemplate: "/users/any"},

		{method: http.MethodPost, path: "/seArcH", valid: true, pathTemplate: "/search"},
		{method: http.MethodPost, path: "/sEarCh/gO", valid: true, pathTemplate: "/search/go"},
		{method: http.MethodPost, path: "/SeArcH/Go1.hTMl", valid: true, pathTemplate: "/search/go1.html"},
		{method: http.MethodPost, path: "/sEaRch/inDEx.hTMl", valid: true, pathTemplate: "/search/index.html"},
		{method: http.MethodPost, path: "/SEARCH/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: http.MethodPost, path: "/SeArCh/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: http.MethodPost, path: "/sEArCH/gophers/Go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: http.MethodPost, path: "/sEArCH/nature/go1.HTML", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: http.MethodPost, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: http.MethodPut, path: "/Src/paris/InValiD", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: http.MethodPut, path: "/SrC/InvaliD", valid: true, pathTemplate: "/src/invalid"},
		{method: http.MethodPut, path: "/SrC1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: http.MethodPut, path: "/SrC1", valid: true, pathTemplate: "/src1"},

		{method: http.MethodPatch, path: "/Signal-R/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: http.MethodPatch, path: "/sIgNaL-r", valid: true, pathTemplate: "/signal-r"},
		{method: http.MethodPatch, path: "/SIGNAL-R/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: http.MethodPatch, path: "/sIGNal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: http.MethodHead, path: "/quERy/unKNown/paGEs", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: http.MethodHead, path: "/QUery/10/amazing/reset/SiNglE", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: http.MethodHead, path: "/QueRy/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: http.MethodHead, path: "/qUERy/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: http.MethodHead, path: "/QueRy/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: http.MethodHead, path: "/qUeRy/uNkNoWn", valid: true, pathTemplate: "/query/unknown"},
		{method: http.MethodHead, path: "/QuerY/UntOld", valid: true, pathTemplate: "/query/untold"},

		{method: http.MethodOptions, path: "/qUestions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: http.MethodOptions, path: "/quEsTioNs", valid: true, pathTemplate: "/questions"},

		{method: http.MethodDelete, path: "/GRAPHQL", valid: true, pathTemplate: "/graphql"},
		{method: http.MethodDelete, path: "/gRapH", valid: true, pathTemplate: "/graph"},
		{method: http.MethodDelete, path: "/grAphQl/cMd", valid: true, pathTemplate: "/graphql/cmd", params: nil},

		{method: http.MethodDelete, path: "/File", valid: true, pathTemplate: "/file", params: nil},
		{method: http.MethodDelete, path: "/fIle/rEmOve", valid: true, pathTemplate: "/file/remove", params: nil},

		{method: http.MethodGet, path: "/heRO-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: http.MethodGet, path: "/HEro-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
	}

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	srv := r.Serve()
	rw := httptest.NewRecorder()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(rw, request)
		}
	}
}

func TestRouterMallocs_ServeHTTP_MixedRoutes(t *testing.T) {
	r := newTestDune()

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

	tt := routerScenario{
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

func BenchmarkRouter_ServeHTTP_ParamRoutes_GETMethod(b *testing.B) {
	r := newTestDune()

	routes := []string{
		"/users/find/:name",
		"/users/:id/delete",
		"/users/groups/:groupId/dump",
		"/users/groups/:groupId/export",
		"/users/:id/update",
		"/search/:q",
		"/search/:q/go",
		"/search/:q/go1.html",
		"/search/:q/:w/index.html",
		"/src/:dest/invalid",
		"/src1/:dest",
		"/signal-r/:cmd",
		"/signal-r/:cmd/reflection",
		"/query/:key",
		"/query/:key/:val",
		"/query/:key/:val/:cmd",
		"/query/:key/:val/:cmd/single",
		"/query/:key/:val/:cmd/single/1",
		"/questions/:index",
		"/graphql/:cmd",
		"/:file",
		"/:file/remove",
		"/hero-:name",
	}

	testers := []string{
		"/users/find/yousuf",
		"/users/john/delete",
		"/users/groups/120/dump",
		"/users/groups/230/export",
		"/users/911/update",
		"/search/ducks",
		"/search/gophers/go",
		"/search/nature/go1.html",
		"/search/generics/types/index.html",
		"/src/paris/invalid",
		"/src1/oslo",
		"/signal-r/push",
		"/signal-r/protos/reflection",
		"/query/911",
		"/query/46/hello",
		"/query/99/sup/update-ttl",
		"/query/10/amazing/reset/single",
		"/query/10/amazing/reset/single/1",
		"/questions/1001",
		"/graphql/stream",
		"/gophers.html",
		"/gophers.html/remove",
		"/hero-goku",
		"/hero-thor",
	}

	for _, route := range routes {
		r.GET(route, fakeHandler())
	}

	requests := make([]*http.Request, len(testers))

	for i, path := range testers {
		requests[i], _ = http.NewRequest("GET", path, nil)
	}

	srv := r.Serve()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_ParamRoutes_RandomMethods(b *testing.B) {
	r := newTestDune()

	routes := map[string]string{
		"/users/find/:name":                http.MethodGet,
		"/users/:id/delete":                http.MethodDelete,
		"/users/groups/:groupId/dump":      http.MethodPost,
		"/users/groups/:groupId/export":    http.MethodPost,
		"/users/:id/update":                http.MethodPut,
		"/search/:q":                       http.MethodGet,
		"/search/:q/go":                    http.MethodGet,
		"/search/:q/go1.html":              http.MethodGet,
		"/search/:q/:w/index.html":         http.MethodOptions,
		"/src/:dest/invalid":               http.MethodGet,
		"/src1/:dest":                      http.MethodGet,
		"/signal-r/:cmd":                   http.MethodGet,
		"/signal-r/:cmd/reflection":        http.MethodGet,
		"/query/:key":                      http.MethodGet,
		"/query/:key/:val":                 http.MethodGet,
		"/query/:key/:val/:cmd":            http.MethodGet,
		"/query/:key/:val/:cmd/single":     http.MethodGet,
		"/query/:key/:val/:cmd/single/1":   http.MethodGet,
		"/questions/:index":                http.MethodHead,
		"/graphql/:cmd":                    http.MethodGet,
		"/search/generics/types/index.css": http.MethodGet,
		"/:file":                           http.MethodPatch,
		"/:file/remove":                    http.MethodPatch,
		"/hero-:name":                      http.MethodGet,
	}

	tests := map[string]string{
		"/users/find/yousuf":                http.MethodGet,
		"/users/john/delete":                http.MethodDelete,
		"/users/groups/120/dump":            http.MethodPost,
		"/users/groups/230/export":          http.MethodPost,
		"/users/911/update":                 http.MethodPut,
		"/search/ducks":                     http.MethodGet,
		"/search/gophers/go":                http.MethodGet,
		"/search/nature/go1.html":           http.MethodGet,
		"/search/generics/types/index.html": http.MethodOptions,
		"/src/paris/invalid":                http.MethodGet,
		"/src1/oslo":                        http.MethodGet,
		"/signal-r/push":                    http.MethodGet,
		"/signal-r/protos/reflection":       http.MethodGet,
		"/query/911":                        http.MethodGet,
		"/query/46/hello":                   http.MethodGet,
		"/query/99/sup/update-ttl":          http.MethodGet,
		"/query/10/amazing/reset/single":    http.MethodGet,
		"/query/10/amazing/reset/single/1":  http.MethodGet,
		"/questions/1001":                   http.MethodHead,
		"/graphql/stream":                   http.MethodGet,
		"/search/generics/types/index.css":  http.MethodGet,
		"/gophers.html":                     http.MethodPatch,
		"/gophers.html/remove":              http.MethodPatch,
		"/hero-goku":                        http.MethodGet,
		"/hero-thor":                        http.MethodGet,
	}

	for route, method := range routes {
		r.Map([]string{method}, route, fakeHandler())
	}

	requests := make([]*http.Request, len(tests))

	i := 0
	for path, method := range tests {
		requests[i], _ = http.NewRequest(method, path, nil)
		i++
	}

	srv := r.Serve()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(nil, request)
		}
	}
}

func TestRouter_BuiltInHTTPMethods(t *testing.T) {
	r := New()

	f := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(t, r.Method == method, fmt.Sprintf("http method > expected: %s, got: %s", r.Method, method))
			return nil
		}
	}

	r.GET("/foo", f(http.MethodGet))
	r.POST("/foo", f(http.MethodPost))
	r.POST("/bar", f(http.MethodPost))
	r.PUT("/bar", f(http.MethodPut))
	r.PATCH("/xyz", f(http.MethodPatch))
	r.HEAD("/abc", f(http.MethodHead))
	r.OPTIONS("/abc", f(http.MethodOptions))

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	r2, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	r3, _ := http.NewRequest(http.MethodPost, "/bar", nil)
	r4, _ := http.NewRequest(http.MethodPut, "/bar", nil)
	r5, _ := http.NewRequest(http.MethodPatch, "/xyz", nil)
	r6, _ := http.NewRequest(http.MethodHead, "/abc", nil)
	r7, _ := http.NewRequest(http.MethodOptions, "/abc", nil)

	rw := httptest.NewRecorder()
	requests := [...]*http.Request{r1, r2, r3, r4, r5, r6, r7}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK, fmt.Sprintf("%p http status > expected: %d, got: %d", req.URL, http.StatusOK, rw.Code))
	}

}

func TestRouter_CustomHTTPMethods(t *testing.T) {
	r := New()

	f := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(t, r.Method == method, fmt.Sprintf("http method > expected: %s, got: %s", method, r.Method))
			return nil
		}
	}

	r.Map([]string{"FOO"}, "/foo", f("FOO"))
	r.Map([]string{"BAR"}, "/bar", f("BAR"))
	r.Map([]string{"XYZ"}, "/xyz", f("XYZ"))
	r.Map([]string{"DUNE"}, "/*loc", f("DUNE"))

	srv := r.Serve()

	r1, _ := http.NewRequest("FOO", "/foo", nil)
	r2, _ := http.NewRequest("BAR", "/bar", nil)
	r3, _ := http.NewRequest("XYZ", "/xyz", nil)
	r4, _ := http.NewRequest("DUNE", "/k8", nil)
	r5, _ := http.NewRequest("DUNE", "/etcd", nil)

	rw := httptest.NewRecorder()
	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK, fmt.Sprintf("%p http status > expected: %d, got: %d", req.URL, http.StatusOK, rw.Code))
	}
}

func TestRouter_HTTPMethods(t *testing.T) {
	r := New()

	f := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(t, r.Method == method, fmt.Sprintf("http method > expected: %s, got: %s", r.Method, method))
			return nil
		}
	}

	r.GET("/foo", f(http.MethodGet))
	r.POST("/foo", f(http.MethodPost))
	r.POST("/bar", f(http.MethodPost))
	r.PUT("/bar", f(http.MethodPut))
	r.PATCH("/xyz", f(http.MethodPatch))
	r.HEAD("/abc", f(http.MethodHead))
	r.OPTIONS("/abc", f(http.MethodOptions))
	r.Map([]string{"FOO"}, "/foo", f("FOO"))
	r.Map([]string{"BAR"}, "/bar", f("BAR"))
	r.Map([]string{"XYZ"}, "/xyz", f("XYZ"))
	r.Map([]string{"DUNE"}, "/*loc", f("DUNE"))

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo", nil)
	r2, _ := http.NewRequest(http.MethodPost, "/foo", nil)
	r3, _ := http.NewRequest(http.MethodPost, "/bar", nil)
	r4, _ := http.NewRequest(http.MethodPut, "/bar", nil)
	r5, _ := http.NewRequest(http.MethodPatch, "/xyz", nil)
	r6, _ := http.NewRequest(http.MethodHead, "/abc", nil)
	r7, _ := http.NewRequest(http.MethodOptions, "/abc", nil)
	r8, _ := http.NewRequest("FOO", "/foo", nil)
	r9, _ := http.NewRequest("BAR", "/bar", nil)
	r10, _ := http.NewRequest("XYZ", "/xyz", nil)
	r11, _ := http.NewRequest("DUNE", "/k8", nil)
	r12, _ := http.NewRequest("DUNE", "/etcd", nil)

	rw := httptest.NewRecorder()
	requests := [...]*http.Request{r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK, fmt.Sprintf("%p http status > expected: %d, got: %d", req.URL, http.StatusOK, rw.Code))
	}

}

func TestRouter_DefaultNotFoundHandler(t *testing.T) {
	r := New()

	r.GET("/foo/foo/foo", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)

	srv.ServeHTTP(rw, req)
	assert(t, rw.Code == http.StatusNotFound, fmt.Sprintf("expected: %d, got: %d", http.StatusNotFound, rw.Code))
}

func TestRouter_CustomNotFoundHandler(t *testing.T) {
	response := "hello from not found handler!"

	r := New()
	r.UseNotFoundHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		_, _ = w.Write([]byte(response))
	})

	r.GET("/foo/foo/foo", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)

	srv.ServeHTTP(rw, req)
	assert(t, rw.Code == http.StatusNotImplemented, fmt.Sprintf("expected: %d, got: %d", http.StatusNotImplemented, rw.Code))
	assert(t, rw.Body.String() == response, fmt.Sprintf("expected: %s, got: %s", response, rw.Body.String()))
}

func TestRouter_TrailingSlash_WithExecute(t *testing.T) {
	r := New()
	r.UseTrailingSlashMatch(WithExecute())

	f := func(path string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(t, path == r.URL.String(), fmt.Sprintf("path assertion > expected: %s, got: %s", path, r.URL.String()))
			return nil
		}
	}

	r.GET("/foo", f("/foo"))
	r.GET("/bar/", f("/bar/"))

	srv := r.Serve()

	rw := httptest.NewRecorder()
	r1, _ := http.NewRequest(http.MethodGet, "/foo/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/bar", nil)

	requests := [...]*http.Request{r1, r2}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK,
			fmt.Sprintf("%s > expected: %d, got: %d", req.URL.String(), http.StatusOK, rw.Code))
	}
}

func TestRouter_TrailingSlash_WithRedirect(t *testing.T) {
	r := New()
	r.UseTrailingSlashMatch(WithRedirect())

	r.GET("/foo", fakeHandler())
	r.GET("/bar/", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()
	r1, _ := http.NewRequest(http.MethodGet, "/foo/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/bar", nil)

	requests := [...]*http.Request{r1, r2}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusMovedPermanently,
			fmt.Sprintf("%s > expected: %d, got: %d", req.URL.String(), http.StatusMovedPermanently, rw.Code))
	}
}

func TestRouter_TrailingSlash_WithRedirectCustom(t *testing.T) {
	r := New()
	r.UseTrailingSlashMatch(WithRedirectCustom(399))

	r.GET("/foo", fakeHandler())
	r.GET("/bar/", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()
	r1, _ := http.NewRequest(http.MethodGet, "/foo/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/bar", nil)

	requests := [...]*http.Request{r1, r2}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == 399,
			fmt.Sprintf("%s > expected: %d, got: %d", req.URL.String(), 399, rw.Code))
	}
}

func TestRouter_SanitizeURL_WithExecute(t *testing.T) {
	r := New()
	r.UseSanitizeURLMatch(WithExecute())

	r.GET("/abc/def", fakeHandler())
	r.GET("/abc/def/ghi", fakeHandler())
	r.GET("/mno", fakeHandler())
	r.GET("/abc/def/jkl", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()

	r1, _ := http.NewRequest(http.MethodGet, "abc/def", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/abc//def//ghi", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/./abc/def", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/abc/def/../../../ghi/jkl/../../../mno", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/abc/def/ghi/../jkl", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK,
			fmt.Sprintf("%s > expected: %d, got: %d", req.URL.String(), http.StatusOK, rw.Code))
	}
}
func TestRouter_SanitizeURL_WithRedirect(t *testing.T) {
	r := New()
	r.UseSanitizeURLMatch(WithRedirect())

	r.GET("/abc/def", fakeHandler())
	r.GET("/abc/def/ghi", fakeHandler())
	r.GET("/mno", fakeHandler())
	r.GET("/abc/def/jkl", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()

	r1, _ := http.NewRequest(http.MethodGet, "abc/def", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/abc//def//ghi", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/./abc/def", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/abc/def/../../../ghi/jkl/../../../mno", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/abc/def/ghi/../jkl", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusMovedPermanently,
			fmt.Sprintf("%s > expected: %d, got: %d", req.URL.String(), http.StatusMovedPermanently, rw.Code))
	}
}

func TestRouter_SanitizeURL_WithRedirectCustom(t *testing.T) {
	r := New()
	r.UseSanitizeURLMatch(WithRedirectCustom(364))

	r.GET("/abc/def", fakeHandler())
	r.GET("/abc/def/ghi", fakeHandler())
	r.GET("/mno", fakeHandler())
	r.GET("/abc/def/jkl", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()

	r1, _ := http.NewRequest(http.MethodGet, "abc/def", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/abc//def//ghi", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/./abc/def", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/abc/def/../../../ghi/jkl/../../../mno", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/abc/def/ghi/../jkl", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == 364,
			fmt.Sprintf("%s > expected: %d, got: %d", req.URL.String(), 364, rw.Code))
	}
}

func TestRouter_MethodNotAllowed_On(t *testing.T) {
	r := New()
	r.UseMethodNotAllowedHandler()

	r.GET("/foo/foo", fakeHandler())
	r.POST("/foo/foo", fakeHandler())
	r.TRACE("/foo/foo", fakeHandler())
	r.CONNECT("/foo/foo", fakeHandler())
	r.Map([]string{"BAR", "XYZ"}, "/foo/foo", fakeHandler())
	r.OPTIONS("/abc", fakeHandler())

	srv := r.Serve()

	expected := []string{http.MethodGet, http.MethodPost, http.MethodTrace, http.MethodConnect, "BAR", "XYZ"}
	sort.Slice(expected, func(i, j int) bool {
		return expected[i] < expected[j]
	})

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/foo/foo", nil)

	srv.ServeHTTP(rw, req)

	allow := strings.Split(rw.Header().Get("Allow"), ", ")
	assert(t, len(expected) == len(allow), fmt.Sprintf("allow list length > expected: %d, got: %d", len(expected), len(allow)))

	sort.Slice(allow, func(i, j int) bool {
		return allow[i] < allow[j]
	})

	for i := 0; i < len(expected); i++ {
		assert(t, expected[i] == allow[i], fmt.Sprintf("allow method > expected: %s, got: %s", expected[i], allow[i]))
	}
}

func TestRouter_MethodNotAllowed_Off(t *testing.T) {
	d := New()

	d.GET("/foo/foo", fakeHandler())
	d.POST("/foo/foo", fakeHandler())
	d.PATCH("/foo/foo", fakeHandler())
	d.PUT("/foo/foo", fakeHandler())
	d.Map([]string{"BAR", "XYZ"}, "/foo/foo", fakeHandler())
	d.OPTIONS("/abc", fakeHandler())

	srv := d.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/foo/foo", nil)

	srv.ServeHTTP(rw, req)
	assert(t, rw.Code == http.StatusNotFound, fmt.Sprintf("http status > expected: %d, got: %d", http.StatusNotFound, rw.Code))

	allow := rw.Header().Get("Allow")
	assert(t, allow == "", fmt.Sprintf("allow header > expected: empty, got: %s", allow))
}

func TestRouter_20ParamsAllocs(t *testing.T) {
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

	allocs := testing.AllocsPerRun(1_000, func() {
		srv.ServeHTTP(rw, req)
		if rw.Code != http.StatusOK {
			panic(fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		}
	})

	assert(t, allocs == 0, fmt.Sprintf("allocs > expected: 0, got: %g", allocs))
}

func TestRouter_200ParamsAllocs(t *testing.T) {
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

	allocs := testing.AllocsPerRun(1_000, func() {
		srv.ServeHTTP(rw, req)
		if rw.Code != http.StatusOK {
			panic(fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		}
	})

	assert(t, allocs == 0, fmt.Sprintf("allocs > expected: 0, got: %g", allocs))
}

func TestRouter_2000ParamsAllocs(t *testing.T) {
	r := New()

	var template strings.Builder
	var path strings.Builder
	var paramKeys []string

	for i := 1; i <= 2_000; i++ {
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

	allocs := testing.AllocsPerRun(1_000, func() {
		srv.ServeHTTP(rw, req)
		if rw.Code != http.StatusOK {
			panic(fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		}
	})

	assert(t, allocs == 0, fmt.Sprintf("allocs > expected: 0, got: %g", allocs))
}

func BenchmarkRouter_2000Params(b *testing.B) {
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
			route.Params.Get(key)
		}
		return nil
	}

	r.GET(template.String(), f)

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(rw, req)
	}
}

func TestRouter_StaticRoutes_EmptyParamsRef(t *testing.T) {
	r := New()

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		assert(t, route.Params == emptyParams, fmt.Sprintf("params ref > expected: %p, got: %p", emptyParams, route.Params))
		return nil
	}

	r.GET("/foo/foo", f)
	r.POST("/foo/foo", f)
	r.HEAD("/bar/index.html", f)
	r.GET("/dune/config.html", f)
	r.GET("/dune/up.yaml", f)

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)
	r2, _ := http.NewRequest(http.MethodPost, "/foo/foo", nil)
	r3, _ := http.NewRequest(http.MethodHead, "/bar/index.html", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/dune/config.html", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/dune/up.yaml", nil)

	rw := httptest.NewRecorder()
	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		srv.ServeHTTP(rw, req)
		assert(t, rw.Code == http.StatusOK, fmt.Sprintf("%p http status > expected: %d, got: %d", req.URL, http.StatusOK, rw.Code))
	}
}

func TestRouter_StaticRoutes_EmptyParamsRef_ConcurrentAccess(t *testing.T) {
	r := New()

	var wg sync.WaitGroup

	paramAccessor := func(ps *Params) {
		ps.appendValue("v") // though the func is not exported, let's ensure we don't add any concurrent unsafe behavior.
		ps.Get("id")
		ps.ForEach(func(k, v string) bool {
			return true
		})
	}

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		defer func() {
			b := recover()
			assert(t, b == nil, fmt.Sprintf("recovery > expected: no panic, got: %v", b))
			wg.Done()
		}()

		assert(t, route.Params == emptyParams, fmt.Sprintf("params ref > expected: %p, got: %p", emptyParams, route.Params))
		paramAccessor(route.Params)
		return nil
	}

	r.GET("/foo/foo", f)
	r.POST("/foo/foo", f)
	r.HEAD("/bar/index.html", f)
	r.GET("/dune/config.html", f)
	r.GET("/dune/up.yaml", f)

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)
	r2, _ := http.NewRequest(http.MethodPost, "/foo/foo", nil)
	r3, _ := http.NewRequest(http.MethodHead, "/bar/index.html", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/dune/config.html", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/dune/up.yaml", nil)

	rw := httptest.NewRecorder()
	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for i := 0; i < 100; i++ {
		for _, req := range requests {
			wg.Add(1)
			go func(req *http.Request) {
				srv.ServeHTTP(rw, req)
				assert(t, rw.Code == http.StatusOK, fmt.Sprintf("%p http status > expected: %d, got: %d", req.URL, http.StatusOK, rw.Code))
			}(req)
		}
	}

	wg.Wait()
}

func TestWithRedirectCustom(t *testing.T) {
	t.Run("in 3XX", func(t *testing.T) {
		statusCode := 333
		rec := panicHandler(func() {
			WithRedirectCustom(statusCode)
		})
		assert(t, rec == nil, fmt.Sprintf("expected not to panic for %d", statusCode))
	})

	t.Run("< 3XX", func(t *testing.T) {
		statusCode := 280
		rec := panicHandler(func() {
			WithRedirectCustom(statusCode)
		})
		assert(t, rec != nil, fmt.Sprintf("expected to panic for %d", statusCode))
	})

	t.Run("> 3XX", func(t *testing.T) {
		statusCode := 420
		rec := panicHandler(func() {
			WithRedirectCustom(statusCode)
		})
		assert(t, rec != nil, fmt.Sprintf("expected to panic for %d", statusCode))
	})
}

func assert(t *testing.T, expectation bool, message string) {
	assertOn(t, true, expectation, message)
}

func assertOn(t *testing.T, condition bool, expectation bool, message string) {
	if condition && !expectation {
		t.Error(message)
	}
}

type recorder struct {
	params *Params
	path   string
}

func (rc *recorder) Handler(path string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		rc.params = route.Params.Copy()
		rc.path = path
		return nil
	}
}

func fakeHandler() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		return nil
	}
}

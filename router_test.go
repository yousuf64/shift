package shift

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
)

type serverTestable interface {
	Name() string
	Test(t *testing.T, srv *Server, rr *routeRecorder)
}

type srvTestTable = []serverTestable

type srvTestItem struct {
	method       string
	path         string
	valid        bool
	pathTemplate string
	params       map[string]string
}

func (st srvTestItem) Name() string {
	return st.path
}

func (st srvTestItem) Test(t *testing.T, srv *Server, rr *routeRecorder) {
	resRec := httptest.NewRecorder()
	req, _ := http.NewRequest(st.method, st.path, nil)

	srv.ServeHTTP(resRec, req)
	ok := resRec.Code == 200

	if st.valid && ok {
		assert(t, rr.path == st.pathTemplate, fmt.Sprintf("%s > path template expected: %s, got: %s", st.path, st.pathTemplate, rr.path))

		for k, v := range st.params {
			actual := rr.params.Get(k)
			assert(t, actual == v, fmt.Sprintf("%s > param '%s' > expected value: '%s', got '%s'", st.path, k, v, actual))
		}

	} else if st.valid && !ok {
		t.Errorf(fmt.Sprintf("%s > expected a handler, but didn't find a handler", st.path))
	} else if !st.valid && ok {
		t.Errorf(fmt.Sprintf("%s > didn't expect a handler, but found a handler with template '%s'", st.path, rr.path))
	}

	rr.Clear()
}

type srvTestItemWithParamsCount struct {
	method       string
	path         string
	valid        bool
	pathTemplate string
	params       map[string]string
	paramsCount  int
}

func (st srvTestItemWithParamsCount) Name() string {
	return st.path
}

func (st srvTestItemWithParamsCount) Test(t *testing.T, srv *Server, rr *routeRecorder) {
	resRec := httptest.NewRecorder()
	req, _ := http.NewRequest(st.method, st.path, nil)

	srv.ServeHTTP(resRec, req)
	ok := resRec.Code == 200

	if st.valid && ok {
		assert(t, rr.path == st.pathTemplate, fmt.Sprintf("%s > path template expected: %s, got: %s", st.path, st.pathTemplate, rr.path))

		gotParamsCount := 0
		if rr.params.internal != nil {
			gotParamsCount = rr.params.Len()
		}
		assert(t, st.paramsCount == gotParamsCount, fmt.Sprintf("%s > params count expected: %d, got: %d", st.path, st.paramsCount, gotParamsCount))

		for k, v := range st.params {
			actual := rr.params.Get(k)
			assert(t, actual == v, fmt.Sprintf("%s > param '%s' > expected value: '%s', got '%s'", st.path, k, v, actual))
		}

	} else if st.valid && !ok {
		t.Errorf(fmt.Sprintf("%s > expected a handler, but didn't find a handler", st.path))
	} else if !st.valid && ok {
		t.Errorf(fmt.Sprintf("%s > didn't expect a handler, but found a handler with template '%s'", st.path, rr.path))
	}

	rr.Clear()
}

type routeRecorder struct {
	params Params
	path   string
}

func (rc *routeRecorder) Handler() HandlerFunc {
	return func(_ http.ResponseWriter, _ *http.Request, route Route) error {
		rc.params = route.Params.Copy()
		rc.path = route.Path
		return nil
	}
}

func (rc *routeRecorder) Clear() {
	rc.params = Params{}
	rc.path = ""
}

func newTestRouter() *Router {
	return New()
}

func testRouter(t *testing.T, srv *Server, rr *routeRecorder, table srvTestTable) {
	for _, st := range table {
		t.Run(st.Name(), func(t *testing.T) {
			st.Test(t, srv, rr)
		})
	}
}

func TestRouter_ServeHTTP_StaticRoutes(t *testing.T) {
	r := newTestRouter()
	rec := &routeRecorder{}

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

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler())
	}

	tt := srvTestTable{
		srvTestItem{method: http.MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		srvTestItem{method: http.MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		srvTestItem{method: http.MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		srvTestItem{method: http.MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		srvTestItem{method: http.MethodGet, path: "/users/all/import", valid: false},
		srvTestItem{method: http.MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		srvTestItem{method: http.MethodGet, path: "/users/911", valid: false},
		srvTestItem{method: http.MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		srvTestItem{method: http.MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		srvTestItem{method: http.MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		srvTestItem{method: http.MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		srvTestItem{method: http.MethodGet, path: "/search/index.html/from-cache", valid: false},
		srvTestItem{method: http.MethodGet, path: "/search/contact.html", valid: false},
		srvTestItem{method: http.MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		srvTestItem{method: http.MethodGet, path: "/src", valid: false},
		srvTestItem{method: http.MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		srvTestItem{method: http.MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		srvTestItem{method: http.MethodGet, path: "/signal-r/connect", valid: false},
		srvTestItem{method: http.MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		srvTestItem{method: http.MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		srvTestItem{method: http.MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		srvTestItem{method: http.MethodGet, path: "/query", valid: false},
		srvTestItem{method: http.MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},
		srvTestItem{method: http.MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		srvTestItem{method: http.MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		srvTestItem{method: http.MethodGet, path: "/graphq", valid: false},
	}

	testRouter(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_ParamRoutes(t *testing.T) {
	r := newTestRouter()
	rec := &routeRecorder{}

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

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler())
	}

	tt := srvTestTable{
		srvTestItem{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		srvTestItem{method: http.MethodGet, path: "/users/find/yousuf/import", valid: false},
		srvTestItem{method: http.MethodDelete, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		srvTestItem{method: http.MethodPut, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		srvTestItem{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		srvTestItem{method: http.MethodGet, path: "/users/groups/230/export/csv", valid: false},
		srvTestItem{method: http.MethodPut, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		srvTestItem{method: http.MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		srvTestItem{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		srvTestItem{method: http.MethodGet, path: "/search/gophers/rust", valid: false},
		srvTestItem{method: http.MethodTrace, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		srvTestItem{method: http.MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},
		srvTestItem{method: http.MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		srvTestItem{method: http.MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		srvTestItem{method: http.MethodGet, path: "/src1/toronto/ontario", valid: false},
		srvTestItem{method: http.MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		srvTestItem{method: http.MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		srvTestItem{method: http.MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		srvTestItem{method: http.MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		srvTestItem{method: http.MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		srvTestItem{method: http.MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		srvTestItem{method: http.MethodGet, path: "/query/10/amazing/reset/single/1", valid: false},
		srvTestItem{method: http.MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		srvTestItem{method: http.MethodPut, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		srvTestItem{method: http.MethodPut, path: "/graphql/stream/tcp", valid: false},
		srvTestItem{method: http.MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		srvTestItem{method: http.MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		srvTestItem{method: http.MethodGet, path: "/gophers.html/fetch", valid: false},
		srvTestItem{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		srvTestItem{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		srvTestItem{method: http.MethodGet, path: "/hero-", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "hero-"}},
	}

	testRouter(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_DifferentParamNames(t *testing.T) {
	r := newTestRouter()
	rr := &routeRecorder{}

	r.GET("/foo/:id", rr.Handler())
	r.GET("/foo/:name/abc", rr.Handler())

	r.GET("/xyz:param", rr.Handler())
	r.GET("/xyz:lang/aaa", rr.Handler())

	r.GET("/www/:filename/:extension", rr.Handler())
	r.GET("/www/:file/:ext/upload", rr.Handler())

	tt := srvTestTable{
		srvTestItem{method: http.MethodGet, path: "/foo/911", valid: true, pathTemplate: "/foo/:id", params: map[string]string{"id": "911"}},
		srvTestItem{method: http.MethodGet, path: "/foo/bar/abc", valid: true, pathTemplate: "/foo/:name/abc", params: map[string]string{"name": "bar"}},

		srvTestItem{method: http.MethodGet, path: "/xyzooo", valid: true, pathTemplate: "/xyz:param", params: map[string]string{"param": "ooo"}},
		srvTestItem{method: http.MethodGet, path: "/xyzgo/aaa", valid: true, pathTemplate: "/xyz:lang/aaa", params: map[string]string{"lang": "go"}},

		srvTestItem{method: http.MethodGet, path: "/www/shift/jpeg", valid: true, pathTemplate: "/www/:filename/:extension", params: map[string]string{"filename": "shift", "extension": "jpeg"}},
		srvTestItem{method: http.MethodGet, path: "/www/meme/gif/upload", valid: true, pathTemplate: "/www/:file/:ext/upload", params: map[string]string{"file": "meme", "ext": "gif"}},
	}

	testRouter(t, r.Serve(), rr, tt)
}

func TestRouter_ServeHTTP_WildcardRoutes(t *testing.T) {
	r := newTestRouter()
	rec := &routeRecorder{}

	paths := map[string]string{
		"/messages/*action":     http.MethodGet,
		"/users/posts/*command": http.MethodGet,
		"/images/*filepath":     http.MethodGet,
		"/hero-*dir":            http.MethodGet,
		"/netflix*abc":          http.MethodGet,
	}

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler())
	}

	tt := srvTestTable{
		srvTestItem{method: http.MethodGet, path: "/messages/publish", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish"}},
		srvTestItem{method: http.MethodGet, path: "/messages/publish/OrderPlaced", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish/OrderPlaced"}},
		srvTestItem{method: http.MethodGet, path: "/messages/", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": ""}},
		srvTestItem{method: http.MethodGet, path: "/messages", valid: false},
		srvTestItem{method: http.MethodGet, path: "/users/posts/", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": ""}},
		srvTestItem{method: http.MethodGet, path: "/users/posts", valid: false},
		srvTestItem{method: http.MethodGet, path: "/users/posts/push", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push"}},
		srvTestItem{method: http.MethodGet, path: "/users/posts/push/911", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push/911"}},
		srvTestItem{method: http.MethodGet, path: "/images/gopher.png", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "gopher.png"}},
		srvTestItem{method: http.MethodGet, path: "/images/", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": ""}},
		srvTestItem{method: http.MethodGet, path: "/images", valid: false},
		srvTestItem{method: http.MethodGet, path: "/images/svg/up-icon", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "svg/up-icon"}},
		srvTestItem{method: http.MethodGet, path: "/hero-dc/batman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/batman.json"}},
		srvTestItem{method: http.MethodGet, path: "/hero-dc/superman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/superman.json"}},
		srvTestItem{method: http.MethodGet, path: "/hero-marvel/loki.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "marvel/loki.json"}},
		srvTestItem{method: http.MethodGet, path: "/hero-", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": ""}},
		srvTestItem{method: http.MethodGet, path: "/hero", valid: false},
		srvTestItem{method: http.MethodGet, path: "/netflix", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": ""}},
		srvTestItem{method: http.MethodGet, path: "/netflix++", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": "++"}},
		srvTestItem{method: http.MethodGet, path: "/netflix/drama/better-call-saul", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": "/drama/better-call-saul"}},
	}

	testRouter(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_MixedRoutes(t *testing.T) {
	r := newTestRouter()
	rec := &routeRecorder{}

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

	for path, meth := range paths {
		r.Map([]string{meth}, path, rec.Handler())
	}

	tt := srvTestTable{
		srvTestItem{method: http.MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		srvTestItem{method: http.MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		srvTestItem{method: http.MethodGet, path: "/users/find/yousuf/import", valid: false},
		srvTestItem{method: http.MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		srvTestItem{method: http.MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		srvTestItem{method: http.MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		srvTestItem{method: http.MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		srvTestItem{method: http.MethodGet, path: "/users/groups/230/export/csv", valid: false},
		srvTestItem{method: http.MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		srvTestItem{method: http.MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		srvTestItem{method: http.MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		srvTestItem{method: http.MethodGet, path: "/users/all/import", valid: false},
		srvTestItem{method: http.MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		srvTestItem{method: http.MethodGet, path: "/users/911", valid: false},

		srvTestItem{method: http.MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		srvTestItem{method: http.MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		srvTestItem{method: http.MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		srvTestItem{method: http.MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		srvTestItem{method: http.MethodGet, path: "/search/index.html/from-cache", valid: false},
		srvTestItem{method: http.MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		srvTestItem{method: http.MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		srvTestItem{method: http.MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		srvTestItem{method: http.MethodGet, path: "/search/gophers/rust", valid: false},
		srvTestItem{method: http.MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		srvTestItem{method: http.MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		srvTestItem{method: http.MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		srvTestItem{method: http.MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		srvTestItem{method: http.MethodGet, path: "/src", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "src"}},
		srvTestItem{method: http.MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		srvTestItem{method: http.MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		srvTestItem{method: http.MethodGet, path: "/src1/toronto/ontario", valid: false},

		srvTestItem{method: http.MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		srvTestItem{method: http.MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		srvTestItem{method: http.MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		srvTestItem{method: http.MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		srvTestItem{method: http.MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		srvTestItem{method: http.MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		srvTestItem{method: http.MethodGet, path: "/query/10/amazing/reset/single/1", valid: false},
		srvTestItem{method: http.MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		srvTestItem{method: http.MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		srvTestItem{method: http.MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		srvTestItem{method: http.MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		srvTestItem{method: http.MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		srvTestItem{method: http.MethodGet, path: "/query", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "query"}},

		srvTestItem{method: http.MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		srvTestItem{method: http.MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},

		srvTestItem{method: http.MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		srvTestItem{method: http.MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		srvTestItem{method: http.MethodGet, path: "/graphq", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "graphq"}},
		srvTestItem{method: http.MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		srvTestItem{method: http.MethodGet, path: "/graphql/stream/tcp", valid: false},

		srvTestItem{method: http.MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		srvTestItem{method: http.MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		srvTestItem{method: http.MethodGet, path: "/gophers.html/fetch", valid: false},

		srvTestItem{method: http.MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		srvTestItem{method: http.MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		srvTestItem{method: http.MethodGet, path: "/hero-", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "hero-"}},
	}

	testRouter(t, r.Serve(), rec, tt)
}

func TestRouter_ServeHTTP_FallbackToParamRoute(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/search/got":           http.MethodGet,
			"/search/:q":            http.MethodGet,
			"/search/:q/go":         http.MethodGet,
			"/search/:q/go/*action": http.MethodGet,
			"/search/:q/*action":    http.MethodGet,
			"/search/*action":       http.MethodGet, // Should never be matched, since it's overridden by a param segment (/search/:q/*action) whose next segment is a wildcard segment.
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gotten", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/got", valid: true, pathTemplate: "/search/got", paramsCount: 0},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gopher", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gopher"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gopher/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gopher"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gok", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gok"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gok/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gok"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gotten/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gotten/goner", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "gotten", "action": "goner"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gotham", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotham"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/got/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "got"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/got/gone", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "gone"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gotham/joker", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "gotham", "action": "joker"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/got/go/pro", valid: true, pathTemplate: "/search/:q/go/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "pro"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/got/go/", valid: true, pathTemplate: "/search/:q/go/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": ""}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/got/apple", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "apple"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/search/go/go/goose":   http.MethodGet,
			"/search/:q":            http.MethodGet,
			"/search/:q/go/goos:x":  http.MethodGet,
			"/search/:q/g:w/goos:x": http.MethodGet,
			"/search/:q/go":         http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gotten", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gox", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gox"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/go/go/goose", valid: true, pathTemplate: "/search/go/go/goose", paramsCount: 0},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/go/go/goosf", valid: true, pathTemplate: "/search/:q/go/goos:x", paramsCount: 2, params: map[string]string{"q": "go", "x": "f"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_FallbackToWildcardRoute(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/search/:q/stop": http.MethodGet,
			"/search/*action": http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/cherry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/cherry/", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/cherry/berry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/berry"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/foo/apple/mango/:fruit": http.MethodGet,
			"/foo/*tag":               http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/foo/apple/orange", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/orange"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/foo/apple/mango/pineapple/another-fruit", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/mango/pineapple/another-fruit"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 3", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/foo/apple/mango/hanna":  http.MethodGet,
			"/foo/apple/mango/:fruit": http.MethodGet,
			"/foo/*tag":               http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/foo/apple/mango/hanna-banana", valid: true, pathTemplate: "/foo/apple/mango/:fruit", paramsCount: 1, params: map[string]string{"fruit": "hanna-banana"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/foo/apple/mango/hanna-banana/watermelon", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/mango/hanna-banana/watermelon"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_RecordParamsOnlyForMatchedPath(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/search":             http.MethodGet,
			"/search/:q/stop":     http.MethodGet,
			"/search/*action":     http.MethodGet,
			"/geo/:lat/:lng/path": http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/cherry/", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/cherry/berry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/berry"}},
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/geo/135/280/path/optimize", valid: false, paramsCount: 0},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/search/go":        http.MethodGet,
			"/search/:var/tail": http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItemWithParamsCount{method: http.MethodGet, path: "/search/gopher", valid: false, paramsCount: 0},
		}

		testRouter(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_Priority(t *testing.T) {
	t.Run("static > wildcard", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/better-call-saul_":        http.MethodGet,
			"/better-call-saul_*season": http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/better-call-saul_", valid: true, pathTemplate: "/better-call-saul_"},
			srvTestItem{method: http.MethodGet, path: "/better-call-saul_6", valid: true, pathTemplate: "/better-call-saul_*season", params: map[string]string{"season": "6"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("param > wildcard", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		paths := map[string]string{
			"/dark_:season": http.MethodGet,
			"/dark_*wc":     http.MethodGet,
		}

		for path, meth := range paths {
			r.Map([]string{meth}, path, rec.Handler())
		}

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/dark_3", valid: true, pathTemplate: "/dark_:season", params: map[string]string{"season": "3"}},
			srvTestItem{method: http.MethodGet, path: "/dark_", valid: true, pathTemplate: "/dark_*wc", params: map[string]string{"wc": ""}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

}

func TestRouter_ServeHTTP_CaseInsensitive(t *testing.T) {
	r := newTestRouter()
	r.UsePathCorrectionMatch(WithExecute())
	rec := &routeRecorder{}

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
		r.Map([]string{meth}, path, rec.Handler())
	}

	tt := srvTestTable{
		srvTestItem{method: http.MethodGet, path: "/users/finD", valid: true, pathTemplate: "/users/find"},
		srvTestItem{method: http.MethodGet, path: "/users/finD/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		srvTestItem{method: http.MethodGet, path: "/users/john/deletE", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		srvTestItem{method: http.MethodGet, path: "/users/911/updatE", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		srvTestItem{method: http.MethodGet, path: "/users/groupS/120/dumP", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		srvTestItem{method: http.MethodGet, path: "/users/groupS/230/exporT", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		srvTestItem{method: http.MethodGet, path: "/users/deletE", valid: true, pathTemplate: "/users/delete"},
		srvTestItem{method: http.MethodGet, path: "/users/alL/dumP", valid: true, pathTemplate: "/users/all/dump"},
		srvTestItem{method: http.MethodGet, path: "/users/alL/exporT", valid: true, pathTemplate: "/users/all/export"},
		srvTestItem{method: http.MethodGet, path: "/users/AnY", valid: true, pathTemplate: "/users/any"},

		srvTestItem{method: http.MethodPost, path: "/seArcH", valid: true, pathTemplate: "/search"},
		srvTestItem{method: http.MethodPost, path: "/sEarCh/gO", valid: true, pathTemplate: "/search/go"},
		srvTestItem{method: http.MethodPost, path: "/SeArcH/Go1.hTMl", valid: true, pathTemplate: "/search/go1.html"},
		srvTestItem{method: http.MethodPost, path: "/sEaRch/inDEx.hTMl", valid: true, pathTemplate: "/search/index.html"},
		srvTestItem{method: http.MethodPost, path: "/SEARCH/contact.html", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "contact.html"}},
		srvTestItem{method: http.MethodPost, path: "/SeArCh/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		srvTestItem{method: http.MethodPost, path: "/sEArCH/gophers/Go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		srvTestItem{method: http.MethodPost, path: "/sEArCH/nature/go1.HTML", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		srvTestItem{method: http.MethodPost, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		srvTestItem{method: http.MethodPut, path: "/Src/paris/InValiD", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		srvTestItem{method: http.MethodPut, path: "/SrC/InvaliD", valid: true, pathTemplate: "/src/invalid"},
		srvTestItem{method: http.MethodPut, path: "/SrC1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		srvTestItem{method: http.MethodPut, path: "/SrC1", valid: true, pathTemplate: "/src1"},

		srvTestItem{method: http.MethodPatch, path: "/Signal-R/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		srvTestItem{method: http.MethodPatch, path: "/sIgNaL-r", valid: true, pathTemplate: "/signal-r"},
		srvTestItem{method: http.MethodPatch, path: "/SIGNAL-R/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		srvTestItem{method: http.MethodPatch, path: "/sIGNal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		srvTestItem{method: http.MethodHead, path: "/quERy/unKNown/paGEs", valid: true, pathTemplate: "/query/unknown/pages"},
		srvTestItem{method: http.MethodHead, path: "/QUery/10/amazing/reset/SiNglE", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		srvTestItem{method: http.MethodHead, path: "/QueRy/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		srvTestItem{method: http.MethodHead, path: "/qUERy/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		srvTestItem{method: http.MethodHead, path: "/QueRy/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		srvTestItem{method: http.MethodHead, path: "/qUeRy/uNkNoWn", valid: true, pathTemplate: "/query/unknown"},
		srvTestItem{method: http.MethodHead, path: "/QuerY/UntOld", valid: true, pathTemplate: "/query/untold"},

		srvTestItem{method: http.MethodOptions, path: "/qUestions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		srvTestItem{method: http.MethodOptions, path: "/quEsTioNs", valid: true, pathTemplate: "/questions"},

		srvTestItem{method: http.MethodDelete, path: "/GRAPHQL", valid: true, pathTemplate: "/graphql"},
		srvTestItem{method: http.MethodDelete, path: "/gRapH", valid: true, pathTemplate: "/graph"},
		srvTestItem{method: http.MethodDelete, path: "/grAphQl/cMd", valid: true, pathTemplate: "/graphql/cmd"},

		srvTestItem{method: http.MethodDelete, path: "/File", valid: true, pathTemplate: "/file"},
		srvTestItem{method: http.MethodDelete, path: "/fIle/rEmOve", valid: true, pathTemplate: "/file/remove"},

		srvTestItem{method: http.MethodGet, path: "/heRO-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		srvTestItem{method: http.MethodGet, path: "/HEro-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
	}

	testRouter(t, r.Serve(), rec, tt)
}

func TestStaticMux_CaseInsensitiveSearch(t *testing.T) {
	r := New()
	r.UsePathCorrectionMatch(WithExecute())

	var subT *testing.T
	f := func(path string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(subT, path == route.Path, fmt.Sprintf("route path expected: %s, got: %s", path, route.Path))
			return nil
		}
	}

	r.GET("/foo", f("/foo"))
	r.GET("/abc/xyz", f("/abc/xyz"))
	r.GET("/go/go/go", f("/go/go/go"))
	r.GET("/bar/ccc/ddd/zzz", f("/bar/ccc/ddd/zzz"))

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/FoO", nil)
	r2, _ := http.NewRequest(http.MethodGet, "../../Foo", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/aBC/xYz", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/gO/GO/go", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/BaR/CcC/ddD/zzZ", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			subT = t
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status expected: %d, got: %d", http.StatusOK, rw.Code))
		})
	}
}

func TestRouter_BuiltInHTTPMethods(t *testing.T) {
	r := newTestRouter()

	var subT *testing.T
	f := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(subT, r.Method == method, fmt.Sprintf("http method > expected: %s, got: %s", r.Method, method))
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

	requests := [...]*http.Request{r1, r2, r3, r4, r5, r6, r7}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			subT = t
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		})
	}

}

func TestRouter_CustomHTTPMethods(t *testing.T) {
	r := newTestRouter()

	var subT *testing.T
	f := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(subT, r.Method == method, fmt.Sprintf("http method > expected: %s, got: %s", method, r.Method))
			return nil
		}
	}

	r.Map([]string{"FOO"}, "/foo", f("FOO"))
	r.Map([]string{"BAR"}, "/bar", f("BAR"))
	r.Map([]string{"XYZ"}, "/xyz", f("XYZ"))
	r.Map([]string{"SHIFT"}, "/*loc", f("SHIFT"))

	srv := r.Serve()

	r1, _ := http.NewRequest("FOO", "/foo", nil)
	r2, _ := http.NewRequest("BAR", "/bar", nil)
	r3, _ := http.NewRequest("XYZ", "/xyz", nil)
	r4, _ := http.NewRequest("SHIFT", "/k8", nil)
	r5, _ := http.NewRequest("SHIFT", "/etcd", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			subT = t
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		})
	}
}

func TestRouter_HTTPMethods(t *testing.T) {
	r := newTestRouter()

	var subT *testing.T
	f := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(subT, r.Method == method, fmt.Sprintf("http method > expected: %s, got: %s", r.Method, method))
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
	r.Map([]string{"SHIFT"}, "/*loc", f("SHIFT"))

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
	r11, _ := http.NewRequest("SHIFT", "/k8", nil)
	r12, _ := http.NewRequest("SHIFT", "/etcd", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5, r6, r7, r8, r9, r10, r11, r12}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			subT = t
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
		})
	}

}

func TestRouter_DefaultNotFoundHandler(t *testing.T) {
	r := newTestRouter()

	r.GET("/foo/foo/foo", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)

	srv.ServeHTTP(rw, req)
	assert(t, rw.Code == http.StatusNotFound, fmt.Sprintf("expected: %d, got: %d", http.StatusNotFound, rw.Code))
}

func TestRouter_CustomNotFoundHandler(t *testing.T) {
	r := newTestRouter()
	response := "hello from not found handler!"

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
	r := newTestRouter()
	r.UseTrailingSlashMatch(WithExecute())

	var subT *testing.T
	f := func(path string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) error {
			assert(subT, path == r.URL.String(), fmt.Sprintf("path assertion > expected: %s, got: %s", path, r.URL.String()))
			return nil
		}
	}

	r.GET("/foo", f("/foo"))
	r.GET("/bar/", f("/bar/"))

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/bar", nil)

	requests := [...]*http.Request{r1, r2}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			subT = t
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("expected: %d, got: %d", http.StatusOK, rw.Code))
		})
	}
}

func TestRouter_TrailingSlash_WithRedirect(t *testing.T) {
	r := newTestRouter()
	r.UseTrailingSlashMatch(WithRedirect())

	r.GET("/foo", fakeHandler())
	r.GET("/bar/", fakeHandler())

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/bar", nil)

	requests := [...]*http.Request{r1, r2}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusMovedPermanently, fmt.Sprintf("expected: %d, got: %d", http.StatusMovedPermanently, rw.Code))
		})
	}
}

func TestRouter_TrailingSlash_WithRedirectCustom(t *testing.T) {
	r := newTestRouter()
	r.UseTrailingSlashMatch(WithRedirectCustom(399))

	r.GET("/foo", fakeHandler())
	r.GET("/bar/", fakeHandler())

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/bar", nil)

	requests := [...]*http.Request{r1, r2}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == 399, fmt.Sprintf("expected: %d, got: %d", 399, rw.Code))
		})
	}
}

func TestRouter_PathCorrection_WithExecute(t *testing.T) {
	r := newTestRouter()
	r.UsePathCorrectionMatch(WithExecute())

	r.GET("/abc/def", fakeHandler())
	r.GET("/abc/def/ghi", fakeHandler())
	r.GET("/mno", fakeHandler())
	r.GET("/abc/def/jkl", fakeHandler())

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "abc/def", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/abc//def//ghi", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/./abc/def", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/abc/def/../../../ghi/jkl/../../../mno", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/abc/def/ghi/../jkl", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("expected: %d, got: %d", http.StatusOK, rw.Code))
		})
	}
}
func TestRouter_PathCorrection_WithRedirect(t *testing.T) {
	r := newTestRouter()
	r.UsePathCorrectionMatch(WithRedirect())

	r.GET("/abc/def", fakeHandler())
	r.GET("/abc/def/ghi", fakeHandler())
	r.GET("/mno", fakeHandler())
	r.GET("/abc/def/jkl", fakeHandler())

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "abc/def", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/abc//def//ghi", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/./abc/def", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/abc/def/../../../ghi/jkl/../../../mno", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/abc/def/ghi/../jkl", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusMovedPermanently,
				fmt.Sprintf("expected: %d, got: %d", http.StatusMovedPermanently, rw.Code))
		})
	}
}

func TestRouter_PathCorrection_WithRedirectCustom(t *testing.T) {
	r := newTestRouter()
	r.UsePathCorrectionMatch(WithRedirectCustom(364))

	r.GET("/abc/def", fakeHandler())
	r.GET("/abc/def/ghi", fakeHandler())
	r.GET("/mno", fakeHandler())
	r.GET("/abc/def/jkl", fakeHandler())

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "abc/def", nil)
	r2, _ := http.NewRequest(http.MethodGet, "/abc//def//ghi", nil)
	r3, _ := http.NewRequest(http.MethodGet, "/./abc/def", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/abc/def/../../../ghi/jkl/../../../mno", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/abc/def/ghi/../jkl", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == 364, fmt.Sprintf("expected: %d, got: %d", 364, rw.Code))
		})
	}
}

func TestRouter_MethodNotAllowed_On(t *testing.T) {
	r := newTestRouter()
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
	r := New()

	r.GET("/foo/foo", fakeHandler())
	r.POST("/foo/foo", fakeHandler())
	r.PATCH("/foo/foo", fakeHandler())
	r.PUT("/foo/foo", fakeHandler())
	r.Map([]string{"BAR", "XYZ"}, "/foo/foo", fakeHandler())
	r.OPTIONS("/abc", fakeHandler())

	srv := r.Serve()

	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/foo/foo", nil)

	srv.ServeHTTP(rw, req)
	assert(t, rw.Code == http.StatusNotFound, fmt.Sprintf("http status > expected: %d, got: %d", http.StatusNotFound, rw.Code))

	allow := rw.Header().Get("Allow")
	assert(t, allow == "", fmt.Sprintf("allow header > expected: empty, got: %s", allow))
}

func TestRouter_StaticRoutes_EmptyParams(t *testing.T) {
	r := newTestRouter()
	var params *Params

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		params = &route.Params
		return nil
	}

	r.GET("/foo/foo", f)
	r.POST("/foo/foo", f)
	r.HEAD("/bar/index.html", f)
	r.GET("/shift/config.html", f)
	r.GET("/shift/up.yaml", f)

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)
	r2, _ := http.NewRequest(http.MethodPost, "/foo/foo", nil)
	r3, _ := http.NewRequest(http.MethodHead, "/bar/index.html", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/shift/config.html", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/shift/up.yaml", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for _, req := range requests {
		t.Run(req.URL.String(), func(t *testing.T) {
			rw := httptest.NewRecorder()
			srv.ServeHTTP(rw, req)
			assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status > want: %d, got: %d", http.StatusOK, rw.Code))
			assert(t, params.internal == nil, fmt.Sprintf("params internal > want: %v, got: %p", nil, params))
			params = nil
		})
	}
}

func TestRouter_StaticRoutes_EmptyParams_ConcurrentAccess(t *testing.T) {
	r := newTestRouter()

	paramAccessor := func(p Params) {
		p.Get("id")
		p.Len()
		p.ForEach(func(k, v string) {
			t.Fatalf("didn't expect the predicate to run")
		})
		p.Slice()
		p.Map()
		p.Copy()
	}

	f := func(w http.ResponseWriter, r *http.Request, route Route) error {
		defer func() {
			b := recover()
			assert(t, b == nil, fmt.Sprintf("%s recovery > expected: no panic, got: %v", r.URL.String(), b))
		}()

		assert(t, route.Params.internal == nil, fmt.Sprintf("%s params internal > expected: %v, got: %p", r.URL.String(), nil, route.Params))
		paramAccessor(route.Params)
		return nil
	}

	r.GET("/foo/foo", f)
	r.POST("/foo/foo", f)
	r.HEAD("/bar/index.html", f)
	r.GET("/shift/config.html", f)
	r.GET("/shift/up.yaml", f)

	srv := r.Serve()

	r1, _ := http.NewRequest(http.MethodGet, "/foo/foo", nil)
	r2, _ := http.NewRequest(http.MethodPost, "/foo/foo", nil)
	r3, _ := http.NewRequest(http.MethodHead, "/bar/index.html", nil)
	r4, _ := http.NewRequest(http.MethodGet, "/shift/config.html", nil)
	r5, _ := http.NewRequest(http.MethodGet, "/shift/up.yaml", nil)

	requests := [...]*http.Request{r1, r2, r3, r4, r5}

	for i := 0; i < 100; i++ {
		for _, req := range requests {
			req := req
			t.Run(fmt.Sprintf("%s%s", req.Method, req.URL.String()), func(t *testing.T) {
				t.Parallel()
				rw := httptest.NewRecorder()
				srv.ServeHTTP(rw, req)
				assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status > expected: %d, got: %d", http.StatusOK, rw.Code))
			})
		}
	}
}

func TestRouter_ServeHTTP_PreventMatchingOnEmptyParamValues_1(t *testing.T) {
	t.Run("full segment param", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		r.GET("/products/:id/reviews", rec.Handler())

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/products//reviews", valid: false},
			srvTestItem{method: http.MethodGet, path: "/products/911/reviews", valid: true, pathTemplate: "/products/:id/reviews", params: map[string]string{"id": "911"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("mid segment param", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		r.GET("/products-:id/reviews", rec.Handler())

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/products-/reviews", valid: false},
			srvTestItem{method: http.MethodGet, path: "/products-911/reviews", valid: true, pathTemplate: "/products-:id/reviews", params: map[string]string{"id": "911"}},
		}

		testRouter(t, r.Serve(), rec, tt)
	})
}

// Refer issue https://github.com/yousuf64/shift/issues/9
func TestRouter_ServeHTTP_PreventMatchingOnEmptyParamValues_2(t *testing.T) {
	t.Run("full segment param", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		r.GET("/posts/:id", rec.Handler())
		r.GET("/:aaa", rec.Handler())
		r.GET("/:abc/:bbb/comments", rec.Handler())

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/posts/400", valid: true, pathTemplate: "/posts/:id", params: map[string]string{"id": "400"}},
			srvTestItem{method: http.MethodGet, path: "//posts////", valid: false},
			srvTestItem{method: http.MethodGet, path: "/posts//", valid: false},
			srvTestItem{method: http.MethodGet, path: "/posts//comments", valid: false},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("mid segment param", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		r.GET("/products:id", rec.Handler())
		r.GET("/products:id/tags", rec.Handler())

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/products101", valid: true, pathTemplate: "/products:id", params: map[string]string{"id": "101"}},
			srvTestItem{method: http.MethodGet, path: "/products101/tags", valid: true, pathTemplate: "/products:id/tags", params: map[string]string{"id": "101"}},
			srvTestItem{method: http.MethodGet, path: "/products/", valid: false},
			srvTestItem{method: http.MethodGet, path: "/products/1", valid: false},
			srvTestItem{method: http.MethodGet, path: "/products//", valid: false},
			srvTestItem{method: http.MethodGet, path: "/products//tags", valid: false},
			srvTestItem{method: http.MethodGet, path: "/products/tags", valid: false},
		}

		testRouter(t, r.Serve(), rec, tt)
	})

	t.Run("root segment param", func(t *testing.T) {
		r := newTestRouter()
		rec := &routeRecorder{}

		r.GET("/:aaa", rec.Handler())

		tt := srvTestTable{
			srvTestItem{method: http.MethodGet, path: "/hello", valid: true, pathTemplate: "/:aaa", params: map[string]string{"aaa": "hello"}},
			srvTestItem{method: http.MethodGet, path: "https://example.com//", valid: false},
			srvTestItem{method: http.MethodGet, path: "https://example.com///", valid: false},
			srvTestItem{method: http.MethodGet, path: "https://example.com///hello", valid: false},
		}

		testRouter(t, r.Serve(), rec, tt)
	})
}

func TestRouter_ServeHTTP_MiddlewarePipeline_ExecutionOrder(t *testing.T) {
	r := newTestRouter()

	mw := func(name string) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request, route Route) error {
				w.Write([]byte(name + "_"))
				return next(w, r, route)
			}
		}
	}

	f := HandlerFunc(func(w http.ResponseWriter, r *http.Request, route Route) error {
		w.Write([]byte(fmt.Sprintf("%s %s", r.Method, route.Path)))
		return nil
	})

	r.Use(mw("aaa"))
	r.Use(mw("bbb"))
	r.Group("/v1", func(g *Group) {
		g.Use(mw("foo"))
		g.GET("", f)
		g.Group("/products", func(g *Group) {
			g.GET("", f)
			g.With(mw("bar")).PUT("", f)
			g.POST("/upload", f)
		})
	})
	r.With(mw("baz"), mw("ccc"), mw("fff")).PATCH("/index", f)
	r.Group("/v2", func(g *Group) {
		g.GET("/products", f)
	})

	testTable := []struct {
		method string
		path   string
		out    string
	}{
		{
			method: http.MethodGet,
			path:   "/v1/products",
			out:    "aaa_bbb_foo_GET /v1/products",
		},
		{
			method: http.MethodPut,
			path:   "/v1/products",
			out:    "aaa_bbb_foo_bar_PUT /v1/products",
		},
		{
			method: http.MethodPost,
			path:   "/v1/products/upload",
			out:    "aaa_bbb_foo_POST /v1/products/upload",
		},
		{
			method: http.MethodPatch,
			path:   "/index",
			out:    "aaa_bbb_baz_ccc_fff_PATCH /index",
		},
		{
			method: http.MethodGet,
			path:   "/v2/products",
			out:    "aaa_bbb_GET /v2/products",
		},
	}

	srv := r.Serve()

	for _, tx := range testTable {
		t.Run(fmt.Sprintf("%s%s", tx.path, tx.path), func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, _ := http.NewRequest(tx.method, tx.path, nil)
			srv.ServeHTTP(rw, req)

			assert(t, rw.Body.String() == tx.out, fmt.Sprintf("expected: %s, got: %s", tx.out, rw.Body.String()))
		})
	}
}

func TestRouter_ServeHTTP_MiddlewarePipeline_ExecutionShortCircuiting(t *testing.T) {
	r := newTestRouter()

	mw := func(name string) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request, route Route) error {
				w.Write([]byte(name + "_"))

				exit := r.URL.Query().Get("exit")
				if exit == name {
					return nil
				}
				return next(w, r, route)
			}
		}
	}

	f := HandlerFunc(func(w http.ResponseWriter, r *http.Request, route Route) error {
		w.Write([]byte(fmt.Sprintf("%s %s", r.Method, route.Path)))
		return nil
	})

	r.Use(mw("@"))
	r.Group("/api", func(g *Group) {
		g.Use(mw("foo"))
		g.GET("/shop", f)
	})
	stack := r.With(mw("bar"))
	stack.GET("/api-versions", f)
	stack.With(mw("zzz")).POST("/api-versions", f)

	testTable := []struct {
		method string
		path   string
		out    string
	}{
		{
			method: http.MethodGet,
			path:   "/api/shop",
			out:    "@_foo_GET /api/shop",
		},
		{
			method: http.MethodGet,
			path:   "/api/shop?exit=@",
			out:    "@_",
		},
		{
			method: http.MethodGet,
			path:   "/api/shop?exit=foo",
			out:    "@_foo_",
		},
		{
			method: http.MethodGet,
			path:   "/api-versions",
			out:    "@_bar_GET /api-versions",
		},
		{
			method: http.MethodGet,
			path:   "/api-versions?exit=@",
			out:    "@_",
		},
		{
			method: http.MethodGet,
			path:   "/api-versions?exit=bar",
			out:    "@_bar_",
		},
		{
			method: http.MethodPost,
			path:   "/api-versions",
			out:    "@_bar_zzz_POST /api-versions",
		},
		{
			method: http.MethodPost,
			path:   "/api-versions?exit=@",
			out:    "@_",
		},
		{
			method: http.MethodPost,
			path:   "/api-versions?exit=bar",
			out:    "@_bar_",
		},
		{
			method: http.MethodPost,
			path:   "/api-versions?exit=zzz",
			out:    "@_bar_zzz_",
		},
	}

	srv := r.Serve()

	for _, tx := range testTable {
		t.Run(fmt.Sprintf("%s%s", tx.path, tx.path), func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, _ := http.NewRequest(tx.method, tx.path, nil)
			srv.ServeHTTP(rw, req)

			assert(t, rw.Body.String() == tx.out, fmt.Sprintf("expected: %s, got: %s", tx.out, rw.Body.String()))
		})
	}
}

func TestRouter_ServeHTTP_MiddlewarePipeline_ExecuteOnlyOnRouteMatch(t *testing.T) {
	r := newTestRouter()

	mw := func(name string) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request, route Route) error {
				w.Write([]byte(name + "_"))
				return next(w, r, route)
			}
		}
	}

	f := HandlerFunc(func(w http.ResponseWriter, r *http.Request, route Route) error {
		w.Write([]byte(fmt.Sprintf("%s %s", r.Method, route.Path)))
		return nil
	})

	r.Use(mw("0"))
	r.With(mw("1")).Group("/schemas", func(g *Group) {
		g.With(mw("2")).Group("/v1", func(g *Group) {
			g.GET("/network-policies", f)
			g.With(mw("3")).POST("/service-accounts", f)
		})
	})

	testTable := []struct {
		method string
		path   string
		out    string
	}{
		{
			method: http.MethodGet,
			path:   "/schemas/v1/network-policies",
			out:    "0_1_2_GET /schemas/v1/network-policies",
		},
		{
			method: http.MethodGet,
			path:   "/schemas/v2/network-policies",
			out:    "404 page not found\n",
		},
		{
			method: http.MethodPost,
			path:   "/schemas/v1/network-policies",
			out:    "404 page not found\n",
		},
		{
			method: http.MethodPost,
			path:   "/schemas/v1/service-accounts",
			out:    "0_1_2_3_POST /schemas/v1/service-accounts",
		},
		{
			method: http.MethodGet,
			path:   "/schemas/v1/service-accounts",
			out:    "404 page not found\n",
		},
		{
			method: http.MethodGet,
			path:   "/schemas/v1/service-accounts/spec",
			out:    "404 page not found\n",
		},
	}

	srv := r.Serve()

	for _, tx := range testTable {
		t.Run(fmt.Sprintf("%s%s", tx.path, tx.path), func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, _ := http.NewRequest(tx.method, tx.path, nil)
			srv.ServeHTTP(rw, req)

			assert(t, rw.Body.String() == tx.out, fmt.Sprintf("expected: %s, got: %s", tx.out, rw.Body.String()))
		})
	}
}

func TestRouter_ServeHTTP_MiddlewarePipeline_IgnoreLateRegistered(t *testing.T) {
	r := newTestRouter()

	mw := func(name string) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request, route Route) error {
				w.Write([]byte(name + "_"))
				return next(w, r, route)
			}
		}
	}

	f := HandlerFunc(func(w http.ResponseWriter, r *http.Request, route Route) error {
		w.Write([]byte(fmt.Sprintf("%s %s", r.Method, route.Path)))
		return nil
	})

	r.Group("/movies", func(g *Group) {
		g.Group("/drama", func(g *Group) {
			g.GET("", f)
			g.Use(mw("bar"))
			g.GET("/:id", f)
		})
	})
	r.GET("/healthz", f)
	r.Use(mw("foo"))

	testTable := []struct {
		method string
		path   string
		out    string
	}{
		{
			method: http.MethodGet,
			path:   "/movies/drama",
			out:    "GET /movies/drama",
		},
		{
			method: http.MethodGet,
			path:   "/movies/drama/:id",
			out:    "bar_GET /movies/drama/:id",
		},
		{
			method: http.MethodGet,
			path:   "/healthz",
			out:    "GET /healthz",
		},
	}

	srv := r.Serve()

	for _, tx := range testTable {
		t.Run(fmt.Sprintf("%s%s", tx.path, tx.path), func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, _ := http.NewRequest(tx.method, tx.path, nil)
			srv.ServeHTTP(rw, req)

			assert(t, rw.Body.String() == tx.out, fmt.Sprintf("expected: %s, got: %s", tx.out, rw.Body.String()))
		})
	}
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

func fakeHandler() HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		return nil
	}
}

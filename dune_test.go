package dune

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRouter2_ServeHTTP(t *testing.T) {
	d := NewDune()

	paths := map[string]string{
		"/users/find":                   MethodGet,
		"/users/find/:name":             MethodGet,
		"/users/:id/delete":             MethodGet,
		"/users/:id/update":             MethodGet,
		"/users/groups/:groupId/dump":   MethodGet,
		"/users/groups/:groupId/export": MethodGet,
		"/users/delete":                 MethodGet,
		"/users/all/dump":               MethodGet,
		"/users/all/export":             MethodGet,
		"/users/any":                    MethodGet,

		"/search":                  MethodGet,
		"/search/go":               MethodGet,
		"/search/go1.html":         MethodGet,
		"/search/index.html":       MethodGet,
		"/search/:q":               MethodGet,
		"/search/:q/go":            MethodGet,
		"/search/:q/go1.html":      MethodGet,
		"/search/:q/:w/index.html": MethodGet,

		"/src/:dest/invalid": MethodGet,
		"/src/invalid":       MethodGet,
		"/src1/:dest":        MethodGet,
		"/src1":              MethodGet,

		"/signal-r/:cmd/reflection": MethodGet,
		"/signal-r":                 MethodGet,
		"/signal-r/:cmd":            MethodGet,

		"/query/unknown/pages":         MethodGet,
		"/query/:key/:val/:cmd/single": MethodGet,
		"/query/:key":                  MethodGet,
		"/query/:key/:val/:cmd":        MethodGet,
		"/query/:key/:val":             MethodGet,
		"/query/unknown":               MethodGet,
		"/query/untold":                MethodGet,

		"/questions/:index": MethodGet,
		"/questions":        MethodGet,

		"/graphql":      MethodGet,
		"/graph":        MethodGet,
		"/graphql/:cmd": MethodGet,

		"/:file":        MethodGet,
		"/:file/remove": MethodGet,

		"/hero-:name": MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, rec.Handler(path))
	}

	tt := routerTestTable1{
		{method: MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		{method: MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		{method: MethodGet, path: "/users/911", valid: false, pathTemplate: ""},

		{method: MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		{method: MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		{method: MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		{method: MethodGet, path: "/src", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "src"}},
		{method: MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		{method: MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		{method: MethodGet, path: "/query", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "query"}},

		{method: MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		{method: MethodGet, path: "/graphq", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "graphq"}},
		{method: MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		{method: MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		{method: MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		{method: MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	r := Compile(d)

	testRouter2_ServeHTTP(t, r, rec, tt)
}

func testRouter2_ServeHTTP(t *testing.T, r *Router2, rec *recorder, table routerTestTable1) {
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

func BenchmarkRouter2_ServeHTTP_StaticRoutes(b *testing.B) {
	d := NewDune()

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
	}

	for _, route := range routes {
		d.Get(route, fakeHandler())
	}

	requests := make([]*http.Request, len(testers))

	for i, path := range testers {
		requests[i], _ = http.NewRequest("GET", path, nil)
	}

	r := Compile(d)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			r.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter2_ServeHTTP_MixedRoutes(b *testing.B) {
	d := NewDune()

	paths := map[string]string{
		"/users/find":                   MethodGet,
		"/users/find/:name":             MethodGet,
		"/users/:id/delete":             MethodGet,
		"/users/:id/update":             MethodGet,
		"/users/groups/:groupId/dump":   MethodGet,
		"/users/groups/:groupId/export": MethodGet,
		"/users/delete":                 MethodGet,
		"/users/all/dump":               MethodGet,
		"/users/all/export":             MethodGet,
		"/users/any":                    MethodGet,

		"/search":                  MethodGet,
		"/search/go":               MethodGet,
		"/search/go1.html":         MethodGet,
		"/search/index.html":       MethodGet,
		"/search/:q":               MethodGet,
		"/search/:q/go":            MethodGet,
		"/search/:q/go1.html":      MethodGet,
		"/search/:q/:w/index.html": MethodGet,

		"/src/:dest/invalid": MethodGet,
		"/src/invalid":       MethodGet,
		"/src1/:dest":        MethodGet,
		"/src1":              MethodGet,

		"/signal-r/:cmd/reflection": MethodGet,
		"/signal-r":                 MethodGet,
		"/signal-r/:cmd":            MethodGet,

		"/query/unknown/pages":         MethodGet,
		"/query/:key/:val/:cmd/single": MethodGet,
		"/query/:key":                  MethodGet,
		"/query/:key/:val/:cmd":        MethodGet,
		"/query/:key/:val":             MethodGet,
		"/query/unknown":               MethodGet,
		"/query/untold":                MethodGet,

		"/questions/:index": MethodGet,
		"/questions":        MethodGet,

		"/graphql":      MethodGet,
		"/graph":        MethodGet,
		"/graphql/:cmd": MethodGet,

		"/:file":        MethodGet,
		"/:file/remove": MethodGet,

		//"/hero-:name": MethodGet,
	}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, fakeHandler())
	}

	tt := routerTestTable1{
		//{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},

		{method: MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},

		{method: MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		//{method: MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		//{method: MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: MethodGet, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: MethodGet, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		//{method: MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		//{method: MethodGet, path: "/users/911", valid: false, pathTemplate: ""},

		{method: MethodGet, path: "/search", valid: true, pathTemplate: "/search"},
		{method: MethodGet, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: MethodGet, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: MethodGet, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		//{method: MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		//{method: MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		//{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		//{method: MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		//{method: MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		//{method: MethodGet, path: "/query", valid: false, pathTemplate: ""},

		{method: MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		//{method: MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		//{method: MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		//{method: MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		//{method: MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		//{method: MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		//{method: MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	r := Compile(d)

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			r.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter2_ServeHTTP_MixedRoutes_2(b *testing.B) {
	d := NewDune()

	paths := map[string]string{
		"/users/find":                   MethodPost,
		"/users/find/:name":             MethodGet,
		"/users/:id/delete":             MethodGet,
		"/users/:id/update":             MethodGet,
		"/users/groups/:groupId/dump":   MethodGet,
		"/users/groups/:groupId/export": MethodGet,
		"/users/delete":                 MethodPost,
		"/users/all/dump":               MethodPost,
		"/users/all/export":             MethodPost,
		"/users/any":                    MethodPost,

		"/search":                  MethodPost,
		"/search/go":               MethodPost,
		"/search/go1.html":         MethodPost,
		"/search/index.html":       MethodPost,
		"/search/:q":               MethodGet,
		"/search/:q/go":            MethodGet,
		"/search/:q/go1.html":      MethodGet,
		"/search/:q/:w/index.html": MethodGet,

		"/src/:dest/invalid": MethodGet,
		"/src/invalid":       MethodPost,
		"/src1/:dest":        MethodGet,
		"/src1":              MethodPost,

		"/signal-r/:cmd/reflection": MethodGet,
		"/signal-r":                 MethodPost,
		"/signal-r/:cmd":            MethodGet,

		"/query/unknown/pages":         MethodPost,
		"/query/:key/:val/:cmd/single": MethodGet,
		"/query/:key":                  MethodGet,
		"/query/:key/:val/:cmd":        MethodGet,
		"/query/:key/:val":             MethodGet,
		"/query/unknown":               MethodPost,
		"/query/untold":                MethodPost,

		"/questions/:index": MethodGet,
		"/questions":        MethodPost,

		"/graphql":      MethodPost,
		"/graph":        MethodPost,
		"/graphql/:cmd": MethodGet,

		"/:file":        MethodGet,
		"/:file/remove": MethodGet,

		//"/hero-:name": MethodGet,
	}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, fakeHandler())
	}

	tt := routerTestTable1{
		//{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},

		{method: MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},

		{method: MethodPost, path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{method: MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		//{method: MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: MethodGet, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: MethodGet, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		//{method: MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: MethodPost, path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{method: MethodPost, path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{method: MethodPost, path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		//{method: MethodGet, path: "/users/all/import", valid: false, pathTemplate: ""},
		{method: MethodPost, path: "/users/any", valid: true, pathTemplate: "/users/any"},
		//{method: MethodGet, path: "/users/911", valid: false, pathTemplate: ""},

		{method: MethodPost, path: "/search", valid: true, pathTemplate: "/search"},
		{method: MethodPost, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: MethodPost, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: MethodPost, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		//{method: MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		//{method: MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: MethodPost, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		//{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: MethodPost, path: "/src1", valid: true, pathTemplate: "/src1"},
		//{method: MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: MethodPost, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: MethodGet, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: MethodPost, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		//{method: MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: MethodPost, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: MethodPost, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		//{method: MethodGet, path: "/query", valid: false, pathTemplate: ""},

		{method: MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: MethodPost, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: MethodPost, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: MethodPost, path: "/graph", valid: true, pathTemplate: "/graph"},
		//{method: MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		//{method: MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		//{method: MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		//{method: MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		//{method: MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		//{method: MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	r := Compile(d)

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			r.ServeHTTP(nil, request)
		}
	}
}

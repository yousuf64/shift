package dune

import (
	"net/http"
	"testing"
)

var fakeHandler1 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

type testItem1 struct {
	path         string
	valid        bool
	pathTemplate string
}

type testItem2 struct {
	path         string
	valid        bool
	pathTemplate string
	params       map[string]string
}

type testTable1 = []testItem1
type testTable2 = []testItem2

func TestStatic(t *testing.T) {
	paths := [...]string{
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
		"/graph",
	}

	tree := newRootNode()

	paramsCount := 0
	for _, path := range paths {
		tree.insert(path, HandlerFunc(fakeHandler1))
		pc := findParamsCount(path)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	tt := testTable1{
		{path: "/users/find", valid: true, pathTemplate: "/users/find"},
		{path: "/users/delete", valid: true, pathTemplate: "/users/delete"},
		{path: "/users/all/dump", valid: true, pathTemplate: "/users/all/dump"},
		{path: "/users/all/import", valid: false, pathTemplate: ""},
		{path: "/users/all/export", valid: true, pathTemplate: "/users/all/export"},
		{path: "/users/any", valid: true, pathTemplate: "/users/any"},
		{path: "/users/911", valid: false, pathTemplate: ""},
		{path: "/search", valid: true, pathTemplate: "/search"},
		{path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		{path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{path: "/search/contact.html", valid: false, pathTemplate: ""},
		{path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		{path: "/src", valid: false, pathTemplate: ""},
		{path: "/src1", valid: true, pathTemplate: "/src1"},
		{path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{path: "/signal-r/connect", valid: false, pathTemplate: ""},
		{path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		{path: "/query", valid: false, pathTemplate: ""},
		{path: "/questions", valid: true, pathTemplate: "/questions"},
		{path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{path: "/graph", valid: true, pathTemplate: "/graph"},
		{path: "/graphq", valid: false, pathTemplate: ""},
	}

	testSearch(t, tree, params, tt)
}

func TestDynamicRoutes(t *testing.T) {
	paths := [...]string{
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
		"/questions/:index",
		"/graphql/:cmd",
		"/:file",
		"/:file/remove",
		"/hero-:name",
	}

	tree := &node{}

	paramsCount := 0
	for _, path := range paths {
		tree.insert(path, HandlerFunc(fakeHandler1))
		pc := findParamsCount(path)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	tt := testTable1{
		{path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name"},
		{path: "/users/find/yousuf/import", valid: false, pathTemplate: ""},
		{path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete"},
		{path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump"},
		{path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export"},
		{path: "/users/groups/230/export/csv", valid: false, pathTemplate: ""},
		{path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update"},
		{path: "/search/ducks", valid: true, pathTemplate: "/search/:q"},
		{path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go"},
		{path: "/search/gophers/rust", valid: false, pathTemplate: ""},
		{path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html"},
		{path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html"},
		{path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid"},
		{path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest"},
		{path: "/src1/toronto/ontario", valid: false, pathTemplate: ""},
		{path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd"},
		{path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection"},
		{path: "/query/911", valid: true, pathTemplate: "/query/:key"},
		{path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val"},
		{path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd"},
		{path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single"},
		{path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: ""},
		{path: "/questions/1001", valid: true, pathTemplate: "/questions/:index"},
		{path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd"},
		{path: "/graphql/stream/tcp", valid: false, pathTemplate: ""},
		{path: "/gophers.html", valid: true, pathTemplate: "/:file"},
		{path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove"},
		{path: "/gophers.html/fetch", valid: false, pathTemplate: ""},
		{path: "/hero-goku", valid: true, pathTemplate: "/hero-:name"},
		{path: "/hero-thor", valid: true, pathTemplate: "/hero-:name"},
		{path: "/hero-", valid: false, pathTemplate: ""},
	}

	testSearch(t, tree, params, tt)
}

func TestDynamicRoutesWithParams(t *testing.T) {
	paths := [...]string{
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
		"/questions/:index",
		"/graphql/:cmd",
		"/:file",
		"/:file/remove",
		"/hero-:name",
	}

	tree := &node{}

	paramsCount := 0
	for _, path := range paths {
		tree.insert(path, HandlerFunc(fakeHandler1))

		pc := findParamsCount(path)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	tt := testTable2{
		{path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},
		{path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},
		{path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		{path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},
		{path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		{path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},
		{path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		{path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	testSearchWithParams(t, tree, params, tt)
}

func TestWildcard(t *testing.T) {
	paths := [...]string{
		"/messages/*action",
		"/users/posts/*command",
		"/images/*filepath",
		"/hero-*dir",
		"/netflix*abc",
	}

	tree := &node{}

	paramsCount := 0
	for _, path := range paths {
		tree.insert(path, HandlerFunc(fakeHandler1))
		pc := findParamsCount(path)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	tt := testTable1{
		{path: "/messages/publish", valid: true, pathTemplate: "/messages/*action"},
		{path: "/messages/publish/OrderPlaced", valid: true, pathTemplate: "/messages/*action"},
		{path: "/messages/", valid: true, pathTemplate: "/messages/*action"},
		{path: "/messages", valid: false, pathTemplate: ""},
		{path: "/users/posts/", valid: true, pathTemplate: "/users/posts/*command"},
		{path: "/users/posts", valid: false, pathTemplate: ""},
		{path: "/users/posts/push", valid: true, pathTemplate: "/users/posts/*command"},
		{path: "/users/posts/push/911", valid: true, pathTemplate: "/users/posts/*command"},
		{path: "/images/gopher.png", valid: true, pathTemplate: "/images/*filepath"},
		{path: "/images/", valid: true, pathTemplate: "/images/*filepath"},
		{path: "/images", valid: false, pathTemplate: ""},
		{path: "/images/svg/up-icon", valid: true, pathTemplate: "/images/*filepath"},
		{path: "/hero-dc/batman.json", valid: true, pathTemplate: "/hero-*dir"},
		{path: "/hero-dc/superman.json", valid: true, pathTemplate: "/hero-*dir"},
		{path: "/hero-marvel/loki.json", valid: true, pathTemplate: "/hero-*dir"},
		{path: "/hero-", valid: true, pathTemplate: "/hero-*dir"},
		{path: "/hero", valid: false, pathTemplate: ""},
		{path: "/netflix", valid: true, pathTemplate: "/netflix*abc"},
		{path: "/netflix++", valid: true, pathTemplate: "/netflix*abc"},
		{path: "/netflix/drama/better-call-saul", valid: true, pathTemplate: "/netflix*abc"},
	}

	testSearch(t, tree, params, tt)
}

func TestWildcardParams(t *testing.T) {
	paths := [...]string{
		"/messages/*action",
		"/users/posts/*command",
		"/images/*filepath",
		"/hero-*dir",
	}

	tree := &node{}

	paramsCount := 0
	for _, path := range paths {
		tree.insert(path, HandlerFunc(fakeHandler1))

		pc := findParamsCount(path)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	tt := testTable2{
		{path: "/messages/", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": ""}}, // todo: fix this issue
		{path: "/messages/publish", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish"}},
		{path: "/messages/publish/OrderPlaced", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish/OrderPlaced"}},
		{path: "/messages", valid: false, pathTemplate: "", params: nil},
		{path: "/users/posts/", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": ""}},
		{path: "/users/posts", valid: false, pathTemplate: "", params: nil},
		{path: "/users/posts/push", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push"}},
		{path: "/users/posts/push/911", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push/911"}},
		{path: "/images/gopher.png", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "gopher.png"}},
		{path: "/images/", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": ""}},
		{path: "/images", valid: false, pathTemplate: "", params: nil},
		{path: "/images/svg/up-icon", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "svg/up-icon"}},
		{path: "/hero-dc/batman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/batman.json"}},
		{path: "/hero-dc/superman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/superman.json"}},
		{path: "/hero-marvel/loki.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "marvel/loki.json"}},
		{path: "/hero-", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": ""}},
		{path: "/hero", valid: false, pathTemplate: "", params: nil},
	}

	testSearchWithParams(t, tree, params, tt)
}

func BenchmarkSimple(b *testing.B) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/cmd/:tool/:sub",
		"/cmd/:tool",
		"/src/*filepath",
		"/search",
		"/search/:query",
		"/files/:dir/*filepath",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/info/:user/public",
		"/info/:user/project/:project",
		"/user_:name",
		"/user_:name/about",
	}

	paramsCount := 0
	for _, route := range routes {
		tree.insert(route, HandlerFunc(fakeHandler1))
		pc := findParamsCount(route)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	match := [...]string{
		"cmd/test/",
		"cmd/test/3",
		"src/any",
		"src/some/file.png",
		"search/",
		"search/someth!ng+in+ünìcodé",
		"files/js/inc/framework.js",
		"doc/go_faq.html",
		"doc/go1.html",
		"info/gordon/public",
		"info/gordon/project/go",
		"user_gopher/go",
		"user_gopher/about",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, s := range match {
			tree.search(s, func() *Params {
				return params
			})
			params.reset()
		}
	}
}

func BenchmarkSimple2(b *testing.B) {
	tree := &node{}

	routes := [...]string{
		//"/users/find/:name",
		//"/users/:id/delete",
		"/users/groups/:groupId/dump",
		"/users/groups/:groupId/export",
		//"/users/:id/update",
		"/search/:q",
		//"/search/:q/go",
		//"/search/:q/go1.html",
		"/search/:q/:w/index.html",
		"/src/:dest/invalid",
		"/src1/:dest",
		"/signal-r/:cmd",
		"/signal-r/:cmd/reflection",
		"/query/:key",
		"/query/:key/:val",
		"/query/:key/:val/:cmd",
		"/query/:key/:val/:cmd/single",
		"/questions/:index",
		"/graphql/:cmd",
		//"/:file",
		//"/:file/remove",
		"/hero-:name",
	}

	paramsCount := 0
	for _, route := range routes {
		tree.insert(route, HandlerFunc(fakeHandler1))
		pc := findParamsCount(route)
		if pc > paramsCount {
			paramsCount = pc
		}
	}

	params := newParams(paramsCount)

	match := [...]string{
		//"/users/find/yousuf",
		//"/users/john/delete",
		"/users/groups/120/dump",
		"/users/groups/230/export",
		//"/users/911/update",
		"/search/ducks",
		//"/search/gophers/go",
		//"/search/nature/go1.html",
		"/search/generics/types/index.html",
		"/src/paris/invalid",
		"/src1/oslo",
		"/signal-r/push",
		"/signal-r/protos/reflection",
		"/query/911",
		"/query/46/hello",
		"/query/99/sup/update-ttl",
		"/query/10/amazing/reset/single",
		//"/query/10/amazing/reset/single/1",
		"/questions/1001",
		"/graphql/stream",
		//"/gophers.html",
		//"/gophers.html/remove",
		"/hero-goku",
		"/hero-thor",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, s := range match {
			tree.search(s, func() *Params {
				return params
			})
			params.reset()
		}
	}
}

func testSearch(t *testing.T, tree *node, params *Params, table testTable1) {
	for _, tx := range table {
		nd, ps := tree.search(tx.path, func() *Params {
			return params
		})
		if tx.valid && (nd == nil || nd.handler == nil) {
			t.Errorf("expected: valid handler, got: no handler: %s", tx.path)
		}
		if !tx.valid && nd != nil && nd.handler != nil {
			t.Errorf("expected: no handler, got: valid handler")
		}
		if tx.pathTemplate != "" && tx.pathTemplate != nd.template {
			t.Errorf("%s expected: %s, got: %s", tx.path, tx.pathTemplate, nd.template)
		}
		if ps != nil {
			ps.reset()
		}
	}
}

func testSearchWithParams(t *testing.T, tree *node, params *Params, table testTable2) {
	for _, tx := range table {
		nd, ps := tree.search(tx.path, func() *Params {
			return newParams(5)
		})
		if tx.valid && (nd == nil || nd.handler == nil) {
			t.Errorf("expected: valid handler, got: no handler: %s", tx.path)
		}
		if !tx.valid && nd != nil && nd.handler != nil {
			t.Errorf("expected: no handler, got: valid handler")
		}
		if tx.pathTemplate != "" && tx.pathTemplate != nd.template {
			t.Errorf("expected: %s, got: %s", tx.pathTemplate, nd.template)
		}
		if tx.params != nil {
			for k, v := range tx.params {
				pv := ps.Get(k)
				if v != pv {
					t.Errorf("params assertion failed. expected: %s, got: %s", v, pv)
				}
			}
		}
		if ps != nil {
			ps.reset()
		}
	}
}

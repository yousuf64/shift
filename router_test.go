package dune

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRouter_ServeHTTP_StaticRoutes(t *testing.T) {
	d := New()

	paths := map[string]string{
		"/users/find":          MethodGet,
		"/users/delete":        MethodGet,
		"/users/all/dump":      MethodGet,
		"/users/all/export":    MethodGet,
		"/users/any":           MethodGet,
		"/search":              MethodGet,
		"/search/go":           MethodGet,
		"/search/go1.html":     MethodGet,
		"/search/index.html":   MethodGet,
		"/src/invalid":         MethodGet,
		"/src1":                MethodGet,
		"/signal-r":            MethodGet,
		"/query/unknown":       MethodGet,
		"/query/unknown/pages": MethodGet,
		"/query/untold":        MethodGet,
		"/questions":           MethodGet,
		"/graphql":             MethodGet,
		"/graph":               MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, rec.Handler(path))
	}

	tt := routerTestTable1{
		{method: MethodGet, path: "/users/find", valid: true, pathTemplate: "/users/find"},
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
		{method: MethodGet, path: "/search/contact.html", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/src1", valid: true, pathTemplate: "/src1"},
		{method: MethodGet, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: MethodGet, path: "/signal-r/connect", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: MethodGet, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: MethodGet, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		{method: MethodGet, path: "/query", valid: false, pathTemplate: ""},
		{method: MethodGet, path: "/questions", valid: true, pathTemplate: "/questions"},
		{method: MethodGet, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: MethodGet, path: "/graph", valid: true, pathTemplate: "/graph"},
		{method: MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
	}

	testRouter_ServeHTTP(t, Compile(d), rec, tt)
}

func TestRouter_ServeHTTP_ParamRoutes(t *testing.T) {
	d := New()

	paths := map[string]string{
		"/users/find/:name":             MethodGet,
		"/users/:id/delete":             MethodDelete,
		"/users/groups/:groupId/dump":   MethodPut,
		"/users/groups/:groupId/export": MethodGet,
		"/users/:id/update":             MethodPut,
		"/search/:q":                    MethodGet,
		"/search/:q/go":                 MethodGet,
		"/search/:q/go1.html":           MethodTrace,
		"/search/:q/:w/index.html":      MethodGet,
		"/src/:dest/invalid":            MethodGet,
		"/src1/:dest":                   MethodGet,
		"/signal-r/:cmd":                MethodGet,
		"/signal-r/:cmd/reflection":     MethodGet,
		"/query/:key":                   MethodGet,
		"/query/:key/:val":              MethodGet,
		"/query/:key/:val/:cmd":         MethodGet,
		"/query/:key/:val/:cmd/single":  MethodGet,
		"/questions/:index":             MethodGet,
		"/graphql/:cmd":                 MethodPut,
		"/:file":                        MethodGet,
		"/:file/remove":                 MethodGet,
		"/hero-:name":                   MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, rec.Handler(path))
	}

	tt := routerTestTable1{
		{method: MethodGet, path: "/users/find/yousuf", valid: true, pathTemplate: "/users/find/:name", params: map[string]string{"name": "yousuf"}},
		{method: MethodGet, path: "/users/find/yousuf/import", valid: false, pathTemplate: "", params: nil},
		{method: MethodDelete, path: "/users/john/delete", valid: true, pathTemplate: "/users/:id/delete", params: map[string]string{"id": "john"}},
		{method: MethodPut, path: "/users/groups/120/dump", valid: true, pathTemplate: "/users/groups/:groupId/dump", params: map[string]string{"groupId": "120"}},
		{method: MethodGet, path: "/users/groups/230/export", valid: true, pathTemplate: "/users/groups/:groupId/export", params: map[string]string{"groupId": "230"}},
		{method: MethodGet, path: "/users/groups/230/export/csv", valid: false, pathTemplate: "", params: nil},
		{method: MethodPut, path: "/users/911/update", valid: true, pathTemplate: "/users/:id/update", params: map[string]string{"id": "911"}},
		{method: MethodGet, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: MethodGet, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		{method: MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: MethodTrace, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: MethodGet, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},
		{method: MethodGet, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: MethodGet, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: MethodGet, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: MethodGet, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: MethodGet, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: MethodGet, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: MethodGet, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		{method: MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: MethodPut, path: "/graphql/stream", valid: true, pathTemplate: "/graphql/:cmd", params: map[string]string{"cmd": "stream"}},
		{method: MethodPut, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/gophers.html", valid: true, pathTemplate: "/:file", params: map[string]string{"file": "gophers.html"}},
		{method: MethodGet, path: "/gophers.html/remove", valid: true, pathTemplate: "/:file/remove", params: map[string]string{"file": "gophers.html"}},
		{method: MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		{method: MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		{method: MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	testRouter_ServeHTTP(t, Compile(d), rec, tt)
}

func TestRouter_ServeHTTP_WildcardRoutes(t *testing.T) {
	d := New()

	paths := map[string]string{
		"/messages/*action":     MethodGet,
		"/users/posts/*command": MethodGet,
		"/images/*filepath":     MethodGet,
		"/hero-*dir":            MethodGet,
		"/netflix*abc":          MethodGet,
	}

	rec := &recorder{}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, rec.Handler(path))
	}

	tt := routerTestTable1{
		{method: MethodGet, path: "/messages/publish", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish"}},
		{method: MethodGet, path: "/messages/publish/OrderPlaced", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": "publish/OrderPlaced"}},
		{method: MethodGet, path: "/messages/", valid: true, pathTemplate: "/messages/*action", params: map[string]string{"action": ""}},
		{method: MethodGet, path: "/messages", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/posts/", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": ""}},
		{method: MethodGet, path: "/users/posts", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/users/posts/push", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push"}},
		{method: MethodGet, path: "/users/posts/push/911", valid: true, pathTemplate: "/users/posts/*command", params: map[string]string{"command": "push/911"}},
		{method: MethodGet, path: "/images/gopher.png", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "gopher.png"}},
		{method: MethodGet, path: "/images/", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": ""}},
		{method: MethodGet, path: "/images", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/images/svg/up-icon", valid: true, pathTemplate: "/images/*filepath", params: map[string]string{"filepath": "svg/up-icon"}},
		{method: MethodGet, path: "/hero-dc/batman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/batman.json"}},
		{method: MethodGet, path: "/hero-dc/superman.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "dc/superman.json"}},
		{method: MethodGet, path: "/hero-marvel/loki.json", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": "marvel/loki.json"}},
		{method: MethodGet, path: "/hero-", valid: true, pathTemplate: "/hero-*dir", params: map[string]string{"dir": ""}},
		{method: MethodGet, path: "/hero", valid: false, pathTemplate: "", params: nil},
		{method: MethodGet, path: "/netflix", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": ""}},
		{method: MethodGet, path: "/netflix++", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": "++"}},
		{method: MethodGet, path: "/netflix/drama/better-call-saul", valid: true, pathTemplate: "/netflix*abc", params: map[string]string{"abc": "/drama/better-call-saul"}},
	}

	testRouter_ServeHTTP(t, Compile(d), rec, tt)
}

func TestRouter_ServeHTTP_MixedRoutes(t *testing.T) {
	d := New()

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

	testRouter_ServeHTTP(t, Compile(d), rec, tt)
}

func TestRouter_ServeHTTP_FallbackToParamRoute(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/search/got":           MethodGet,
			"/search/:q":            MethodGet,
			"/search/:q/go":         MethodGet,
			"/search/:q/go/*action": MethodGet,
			"/search/:q/*action":    MethodGet,
			"/search/*action":       MethodGet, // Should never be matched, since it's overridden by a param segment (/search/:q/*action) whose next segment is a wildcard segment.
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/search/gotten", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			{method: MethodGet, path: "/search/got", valid: true, pathTemplate: "/search/got", paramsCount: 0, params: nil},
			{method: MethodGet, path: "/search/gopher", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gopher"}},
			{method: MethodGet, path: "/search/gopher/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gopher"}},
			{method: MethodGet, path: "/search/gok", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gok"}},
			{method: MethodGet, path: "/search/gok/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gok"}},
			{method: MethodGet, path: "/search/gotten/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			{method: MethodGet, path: "/search/gotten/goner", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "gotten", "action": "goner"}},
			{method: MethodGet, path: "/search/gotham", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotham"}},
			{method: MethodGet, path: "/search/got/go", valid: true, pathTemplate: "/search/:q/go", paramsCount: 1, params: map[string]string{"q": "got"}},
			{method: MethodGet, path: "/search/got/gone", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "gone"}},
			{method: MethodGet, path: "/search/gotham/joker", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "gotham", "action": "joker"}},
			{method: MethodGet, path: "/search/got/go/pro", valid: true, pathTemplate: "/search/:q/go/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "pro"}},
			{method: MethodGet, path: "/search/got/go/", valid: true, pathTemplate: "/search/:q/go/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": ""}},
			{method: MethodGet, path: "/search/got/apple", valid: true, pathTemplate: "/search/:q/*action", paramsCount: 2, params: map[string]string{"q": "got", "action": "apple"}},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/search/go/go/goose":   MethodGet,
			"/search/:q":            MethodGet,
			"/search/:q/go/goos:x":  MethodGet,
			"/search/:q/g:w/goos:x": MethodGet,
			"/search/:q/go":         MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/search/gotten", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gotten"}},
			{method: MethodGet, path: "/search/gox", valid: true, pathTemplate: "/search/:q", paramsCount: 1, params: map[string]string{"q": "gox"}},
			{method: MethodGet, path: "/search/go/go/goose", valid: true, pathTemplate: "/search/go/go/goose", paramsCount: 0, params: nil},
			{method: MethodGet, path: "/search/go/go/goosf", valid: true, pathTemplate: "/search/:q/go/goos:x", paramsCount: 2, params: map[string]string{"q": "go", "x": "f"}},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)
	})
}

func TestRouter_ServeHTTP_FallbackToWildcardRoute(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/search/:q/stop": MethodGet,
			"/search/*action": MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/search/cherry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry"}},
			{method: MethodGet, path: "/search/cherry/", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/"}},
			{method: MethodGet, path: "/search/cherry/berry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/berry"}},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/foo/apple/mango/:fruit": MethodGet,
			"/foo/*tag":               MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/foo/apple/orange", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/orange"}},
			{method: MethodGet, path: "/foo/apple/mango/pineapple/another-fruit", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/mango/pineapple/another-fruit"}},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)

	})

	t.Run("scenario 3", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/foo/apple/mango/hanna":  MethodGet,
			"/foo/apple/mango/:fruit": MethodGet,
			"/foo/*tag":               MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/foo/apple/mango/hanna-banana", valid: true, pathTemplate: "/foo/apple/mango/:fruit", paramsCount: 1, params: map[string]string{"fruit": "hanna-banana"}},
			{method: MethodGet, path: "/foo/apple/mango/hanna-banana/watermelon", valid: true, pathTemplate: "/foo/*tag", paramsCount: 1, params: map[string]string{"tag": "apple/mango/hanna-banana/watermelon"}},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)
	})
}

func TestRouter_ServeHTTP_RecordParamsOnlyForMatchedPath(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/search":             MethodGet,
			"/search/:q/stop":     MethodGet,
			"/search/*action":     MethodGet,
			"/geo/:lat/:lng/path": MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/search/cherry/", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/"}},
			{method: MethodGet, path: "/search/cherry/berry", valid: true, pathTemplate: "/search/*action", paramsCount: 1, params: map[string]string{"action": "cherry/berry"}},
			{method: MethodGet, path: "/geo/135/280/path/optimize", valid: false, pathTemplate: "", paramsCount: 0, params: nil},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)
	})

	t.Run("scenario 2", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/search/go":        MethodGet,
			"/search/:var/tail": MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable2{
			{method: MethodGet, path: "/search/gopher", valid: false, pathTemplate: "", paramsCount: 0, params: nil},
		}

		testRouter_ServeHTTP_2(t, Compile(d), rec, tt)
	})
}

func TestRouter_ServeHTTP_Priority(t *testing.T) {
	t.Run("static > wildcard", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/better-call-saul_":        MethodGet,
			"/better-call-saul_*season": MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable1{
			{method: MethodGet, path: "/better-call-saul_", valid: true, pathTemplate: "/better-call-saul_", params: nil},
			{method: MethodGet, path: "/better-call-saul_6", valid: true, pathTemplate: "/better-call-saul_*season", params: map[string]string{"season": "6"}},
		}

		testRouter_ServeHTTP(t, Compile(d), rec, tt)
	})

	t.Run("param > wildcard", func(t *testing.T) {
		d := New()

		paths := map[string]string{
			"/dark_:season": MethodGet,
			"/dark_*wc":     MethodGet,
		}

		rec := &recorder{}

		for path, meth := range paths {
			d.Map(Methods{meth}, path, rec.Handler(path))
		}

		tt := routerTestTable1{
			{method: MethodGet, path: "/dark_3", valid: true, pathTemplate: "/dark_:season", params: map[string]string{"season": "3"}},
			{method: MethodGet, path: "/dark_", valid: true, pathTemplate: "/dark_*wc", params: map[string]string{"wc": ""}},
		}

		testRouter_ServeHTTP(t, Compile(d), rec, tt)
	})

}

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

type routerTestTable1 = []routerTestItem1

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

func testRouter_ServeHTTP(t *testing.T, r *Router, rec *recorder, table routerTestTable1) {
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

func testRouter_ServeHTTP_2(t *testing.T, r *Router, rec *recorder, table routerTestTable2) {
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
			gotParamsCount = len(rec.params.params)
		}
		assert(t, tx.paramsCount == gotParamsCount, fmt.Sprintf("%s > params count expected: %d, got: %d", tx.path, tx.paramsCount, gotParamsCount))

		if len(tx.params) > 0 {
			for k, v := range tx.params {
				actual := rec.params.Get(k)
				assert(t, actual == v, fmt.Sprintf("%s > param '%s' > expected value: '%s', got '%s'", tx.path, k, v, actual))
			}
		}
	}
}

func BenchmarkRouter_ServeHTTP_StaticRoutes(b *testing.B) {
	d := New()

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

func BenchmarkRouter_ServeHTTP_MixedRoutes_S(b *testing.B) {
	d := New()

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

func BenchmarkRouter_ServeHTTP_MixedRoutes_X_1(b *testing.B) {
	d := New()

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

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
	}

	r := Compile(d)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			r.ServeHTTP(nil, request)
		}
	}
}

func BenchmarkRouter_ServeHTTP_MixedRoutes_X_2(b *testing.B) {
	d := New()

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

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
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

func BenchmarkRouter_ServeHTTP_MixedRoutes_X_3(b *testing.B) {
	d := New()

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

		"/search":                  MethodPost,
		"/search/go":               MethodPost,
		"/search/go1.html":         MethodPost,
		"/search/index.html":       MethodPost,
		"/search/:q":               MethodPost,
		"/search/:q/go":            MethodPost,
		"/search/:q/go1.html":      MethodPost,
		"/search/:q/:w/index.html": MethodPost,

		"/src/:dest/invalid": MethodPut,
		"/src/invalid":       MethodPut,
		"/src1/:dest":        MethodPut,
		"/src1":              MethodPut,

		"/signal-r/:cmd/reflection": MethodPatch,
		"/signal-r":                 MethodPatch,
		"/signal-r/:cmd":            MethodPatch,

		"/query/unknown/pages":         MethodHead,
		"/query/:key/:val/:cmd/single": MethodHead,
		"/query/:key":                  MethodHead,
		"/query/:key/:val/:cmd":        MethodHead,
		"/query/:key/:val":             MethodHead,
		"/query/unknown":               MethodHead,
		"/query/untold":                MethodHead,

		"/questions/:index": MethodConnect,
		"/questions":        MethodConnect,

		"/graphql":     MethodDelete,
		"/graph":       MethodDelete,
		"/graphql/cmd": MethodDelete,

		"/file":        MethodDelete,
		"/file/remove": MethodDelete,

		//"/hero-:name": MethodGet,
	}

	for path, meth := range paths {
		d.Map(Methods{meth}, path, fakeHandler())
	}

	tt := routerTestTable1{
		//{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},

		{method: MethodPost, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},

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

		{method: MethodPost, path: "/search", valid: true, pathTemplate: "/search"},
		{method: MethodPost, path: "/search/go", valid: true, pathTemplate: "/search/go"},
		{method: MethodPost, path: "/search/go1.html", valid: true, pathTemplate: "/search/go1.html"},
		{method: MethodPost, path: "/search/index.html", valid: true, pathTemplate: "/search/index.html"},
		//{method: MethodGet, path: "/search/index.html/from-cache", valid: false, pathTemplate: ""},
		{method: MethodPost, path: "/search/contact.html", valid: true, pathTemplate: "/search/:q"},
		{method: MethodPost, path: "/search/ducks", valid: true, pathTemplate: "/search/:q", params: map[string]string{"q": "ducks"}},
		{method: MethodPost, path: "/search/gophers/go", valid: true, pathTemplate: "/search/:q/go", params: map[string]string{"q": "gophers"}},
		//{method: MethodGet, path: "/search/gophers/rust", valid: false, pathTemplate: "", params: nil},
		{method: MethodPost, path: "/search/nature/go1.html", valid: true, pathTemplate: "/search/:q/go1.html", params: map[string]string{"q": "nature"}},
		{method: MethodPost, path: "/search/generics/types/index.html", valid: true, pathTemplate: "/search/:q/:w/index.html", params: map[string]string{"q": "generics", "w": "types"}},

		{method: MethodPut, path: "/src/paris/invalid", valid: true, pathTemplate: "/src/:dest/invalid", params: map[string]string{"dest": "paris"}},
		{method: MethodPut, path: "/src/invalid", valid: true, pathTemplate: "/src/invalid"},
		//{method: MethodGet, path: "/src", valid: false, pathTemplate: ""},
		{method: MethodPut, path: "/src1/oslo", valid: true, pathTemplate: "/src1/:dest", params: map[string]string{"dest": "oslo"}},
		{method: MethodPut, path: "/src1", valid: true, pathTemplate: "/src1"},
		//{method: MethodGet, path: "/src1/toronto/ontario", valid: false, pathTemplate: "", params: nil},

		{method: MethodPatch, path: "/signal-r/protos/reflection", valid: true, pathTemplate: "/signal-r/:cmd/reflection", params: map[string]string{"cmd": "protos"}},
		{method: MethodPatch, path: "/signal-r", valid: true, pathTemplate: "/signal-r"},
		{method: MethodPatch, path: "/signal-r/push", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "push"}},
		{method: MethodPatch, path: "/signal-r/connect", valid: true, pathTemplate: "/signal-r/:cmd", params: map[string]string{"cmd": "connect"}},

		{method: MethodHead, path: "/query/unknown/pages", valid: true, pathTemplate: "/query/unknown/pages"},
		{method: MethodHead, path: "/query/10/amazing/reset/single", valid: true, pathTemplate: "/query/:key/:val/:cmd/single", params: map[string]string{"key": "10", "val": "amazing", "cmd": "reset"}},
		//{method: MethodGet, path: "/query/10/amazing/reset/single/1", valid: false, pathTemplate: "", params: nil},
		{method: MethodHead, path: "/query/911", valid: true, pathTemplate: "/query/:key", params: map[string]string{"key": "911"}},
		{method: MethodHead, path: "/query/99/sup/update-ttl", valid: true, pathTemplate: "/query/:key/:val/:cmd", params: map[string]string{"key": "99", "val": "sup", "cmd": "update-ttl"}},
		{method: MethodHead, path: "/query/46/hello", valid: true, pathTemplate: "/query/:key/:val", params: map[string]string{"key": "46", "val": "hello"}},
		{method: MethodHead, path: "/query/unknown", valid: true, pathTemplate: "/query/unknown"},
		{method: MethodHead, path: "/query/untold", valid: true, pathTemplate: "/query/untold"},
		//{method: MethodGet, path: "/query", valid: false, pathTemplate: ""},

		{method: MethodConnect, path: "/questions/1001", valid: true, pathTemplate: "/questions/:index", params: map[string]string{"index": "1001"}},
		{method: MethodConnect, path: "/questions", valid: true, pathTemplate: "/questions"},

		{method: MethodDelete, path: "/graphql", valid: true, pathTemplate: "/graphql"},
		{method: MethodDelete, path: "/graph", valid: true, pathTemplate: "/graph"},
		//{method: MethodGet, path: "/graphq", valid: false, pathTemplate: ""},
		{method: MethodDelete, path: "/graphql/cmd", valid: true, pathTemplate: "/graphql/cmd", params: nil},
		//{method: MethodGet, path: "/graphql/stream/tcp", valid: false, pathTemplate: "", params: nil},

		{method: MethodDelete, path: "/file", valid: true, pathTemplate: "/file", params: nil},
		{method: MethodDelete, path: "/file/remove", valid: true, pathTemplate: "/file/remove", params: nil},
		//{method: MethodGet, path: "/gophers.html/fetch", valid: false, pathTemplate: "", params: nil},

		//{method: MethodGet, path: "/hero-goku", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "goku"}},
		//{method: MethodGet, path: "/hero-thor", valid: true, pathTemplate: "/hero-:name", params: map[string]string{"name": "thor"}},
		//{method: MethodGet, path: "/hero-", valid: false, pathTemplate: "", params: nil},
	}

	requests := make([]*http.Request, len(tt))

	for i, tx := range tt {
		requests[i], _ = http.NewRequest(tx.method, tx.path, nil)
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

func BenchmarkRouter_ServeHTTP_ParamRoutes_GETMethod(b *testing.B) {
	d := New()

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
		//"/query/:key/:val/:cmd/single/1",
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
		//"/query/10/amazing/reset/single/1",
		"/questions/1001",
		"/graphql/stream",
		"/gophers.html",
		"/gophers.html/remove",
		"/hero-goku",
		"/hero-thor",
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

func BenchmarkRouter_ServeHTTP_ParamRoutes_RandomMethods(b *testing.B) {
	d := New()

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
		d.Map(Methods{method}, route, fakeHandler())
	}

	requests := make([]*http.Request, len(tests))

	i := 0
	for path, method := range tests {
		requests[i], _ = http.NewRequest(method, path, nil)
		i++
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

func (rc *recorder) Handler(path string) Handler {
	return func(w http.ResponseWriter, r *http.Request, p *Params) {
		rc.params = p
		rc.path = path
	}
}

func fakeHandler() Handler {
	return func(w http.ResponseWriter, r *http.Request, p *Params) {

	}
}

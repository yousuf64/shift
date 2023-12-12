package shift

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type mockRW struct {
	headers http.Header
}

func newMockRW() *mockRW {
	return &mockRW{
		http.Header{},
	}
}

func (m *mockRW) Header() (h http.Header) {
	return m.headers
}

func (m *mockRW) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockRW) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockRW) WriteHeader(int) {}

func BenchmarkRouter_2000Params(b *testing.B) {
	r := newTestRouter()

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
			route.Params.Get(key)
		}
		return nil
	}

	r.GET(template.String(), f)

	srv := r.Serve()

	req, _ := http.NewRequest(http.MethodGet, path.String(), nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(nil, req)
	}
}

func BenchmarkRouter_ServeHTTP_StaticRoutes(b *testing.B) {
	r := newTestRouter()

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

	tests := []string{
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

	requests := make([]*http.Request, len(tests))

	for i, path := range tests {
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

func BenchmarkRouter_ServeHTTP_ParamRoutes_GET(b *testing.B) {
	r := newTestRouter()

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

	tests := []string{
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

	requests := make([]*http.Request, len(tests))

	for i, path := range tests {
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
	r := newTestRouter()

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

	requests := make([]*http.Request, 0, len(tests))

	for path, method := range tests {
		req, _ := http.NewRequest(method, path, nil)
		requests = append(requests, req)
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

func BenchmarkRouter_ServeHTTP_MixedRoutes_1(b *testing.B) {
	r := newTestRouter()

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

	tests := []string{
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

	requests := make([]*http.Request, len(tests))

	for i, path := range tests {
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

func BenchmarkRouter_ServeHTTP_MixedRoutes_2(b *testing.B) {
	r := newTestRouter()

	routes := map[string]string{
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

	for path, meth := range routes {
		r.Map([]string{meth}, path, fakeHandler())
	}

	tests := map[string]string{
		"/users/find":        http.MethodGet,
		"/users/find/yousuf": http.MethodGet,
		//"/users/find/yousuf/import": http.MethodGet,
		"/users/john/delete":       http.MethodGet,
		"/users/911/update":        http.MethodGet,
		"/users/groups/120/dump":   http.MethodGet,
		"/users/groups/230/export": http.MethodGet,
		//"/users/groups/230/export/csv": http.MethodGet,
		"/users/delete":     http.MethodGet,
		"/users/all/dump":   http.MethodGet,
		"/users/all/export": http.MethodGet,
		//"/users/all/import": http.MethodGet,
		"/users/any": http.MethodGet,
		//"/users/911": http.MethodGet,

		"/search":            http.MethodGet,
		"/search/go":         http.MethodGet,
		"/search/go1.html":   http.MethodGet,
		"/search/index.html": http.MethodGet,
		//"/search/index.html/from-cache": http.MethodGet,
		"/search/contact.html": http.MethodGet,
		"/search/ducks":        http.MethodGet,
		"/search/gophers/go":   http.MethodGet,
		//"/search/gophers/rust":              http.MethodGet,
		"/search/nature/go1.html":           http.MethodGet,
		"/search/generics/types/index.html": http.MethodGet,

		"/src/paris/invalid": http.MethodGet,
		"/src/invalid":       http.MethodGet,
		//"/src":               http.MethodGet,
		"/src1/oslo": http.MethodGet,
		"/src1":      http.MethodGet,
		//"/src1/toronto/ontario": http.MethodGet,

		"/signal-r/protos/reflection": http.MethodGet,
		"/signal-r":                   http.MethodGet,
		"/signal-r/push":              http.MethodGet,
		"/signal-r/connect":           http.MethodGet,

		"/query/unknown/pages":           http.MethodGet,
		"/query/10/amazing/reset/single": http.MethodGet,
		//"/query/10/amazing/reset/single/1": http.MethodGet,
		"/query/911":               http.MethodGet,
		"/query/99/sup/update-ttl": http.MethodGet,
		"/query/46/hello":          http.MethodGet,
		"/query/unknown":           http.MethodGet,
		"/query/untold":            http.MethodGet,
		//"/query":                   http.MethodGet,

		"/questions/1001": http.MethodGet,
		"/questions":      http.MethodGet,

		"/graphql": http.MethodGet,
		"/graph":   http.MethodGet,
		//"/graphq":         http.MethodGet,
		"/graphql/stream": http.MethodGet,
		//"/graphql/stream/tcp": http.MethodGet,

		"/gophers.html":        http.MethodGet,
		"/gophers.html/remove": http.MethodGet,
		//"/gophers.html/fetch":  http.MethodGet,

		//"/hero-goku": http.MethodGet,
		//"/hero-thor": http.MethodGet,
		//"/hero-":     http.MethodGet,
	}

	requests := make([]*http.Request, 0, len(tests))

	for path, method := range tests {
		req, _ := http.NewRequest(method, path, nil)
		requests = append(requests, req)
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

func BenchmarkRouter_ServeHTTP_MixedRoutes_3(b *testing.B) {
	r := newTestRouter()

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

	tests := map[string]string{
		"/users/find":        http.MethodPost,
		"/users/find/yousuf": http.MethodGet,
		//"/users/find/yousuf/import": http.MethodGet,
		"/users/john/delete":       http.MethodGet,
		"/users/911/update":        http.MethodGet,
		"/users/groups/120/dump":   http.MethodGet,
		"/users/groups/230/export": http.MethodGet,
		//"/users/groups/230/export/csv": http.MethodGet,
		"/users/delete":     http.MethodPost,
		"/users/all/dump":   http.MethodPost,
		"/users/all/export": http.MethodPost,
		//"/users/all/import": http.MethodGet,
		"/users/any": http.MethodPost,
		// "/users/911": http.MethodGet,

		"/search":            http.MethodPost,
		"/search/go":         http.MethodPost,
		"/search/go1.html":   http.MethodPost,
		"/search/index.html": http.MethodPost,
		//"/search/index.html/from-cache": http.MethodGet,
		"/search/contact.html": http.MethodGet,
		"/search/ducks":        http.MethodGet,
		"/search/gophers/go":   http.MethodGet,
		// "/search/gophers/rust": http.MethodGet,
		"/search/nature/go1.html":           http.MethodGet,
		"/search/generics/types/index.html": http.MethodGet,

		"/src/paris/invalid": http.MethodGet,
		"/src/invalid":       http.MethodPost,
		// "/src": http.MethodGet,
		"/src1/oslo": http.MethodGet,
		"/src1":      http.MethodPost,
		// "/src1/toronto/ontario": http.MethodGet,

		"/signal-r/protos/reflection": http.MethodGet,
		"/signal-r":                   http.MethodPost,
		"/signal-r/push":              http.MethodGet,
		"/signal-r/connect":           http.MethodGet,

		"/query/unknown/pages":           http.MethodPost,
		"/query/10/amazing/reset/single": http.MethodGet,
		// "/query/10/amazing/reset/single/1": http.MethodGet,
		"/query/911":               http.MethodGet,
		"/query/99/sup/update-ttl": http.MethodGet,
		"/query/46/hello":          http.MethodGet,
		"/query/unknown":           http.MethodPost,
		"/query/untold":            http.MethodPost,
		// "/query": http.MethodGet,

		"/questions/1001": http.MethodGet,
		"/questions":      http.MethodPost,

		"/graphql": http.MethodPost,
		"/graph":   http.MethodPost,
		// "/graphq": http.MethodGet,
		"/graphql/stream": http.MethodGet,
		// "/graphql/stream/tcp": http.MethodGet,

		"/gophers.html":        http.MethodGet,
		"/gophers.html/remove": http.MethodGet,
		// "/gophers.html/fetch": http.MethodGet,

		// "/hero-goku": http.MethodGet,
		// "/hero-thor": http.MethodGet,
		// "/hero-": http.MethodGet,
	}

	requests := make([]*http.Request, 0, len(tests))

	for path, method := range tests {
		req, _ := http.NewRequest(method, path, nil)
		requests = append(requests, req)
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

func BenchmarkRouter_ServeHTTP_MixedRoutes_4(b *testing.B) {
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

	tests := map[string]string{
		"/users/find":              http.MethodGet,
		"/users/find/yousuf":       http.MethodGet,
		"/users/john/delete":       http.MethodGet,
		"/users/911/update":        http.MethodGet,
		"/users/groups/120/dump":   http.MethodGet,
		"/users/groups/230/export": http.MethodGet,
		"/users/delete":            http.MethodGet,
		"/users/all/dump":          http.MethodGet,
		"/users/all/export":        http.MethodGet,
		"/users/any":               http.MethodGet,

		"/search":                           http.MethodPost,
		"/search/go":                        http.MethodPost,
		"/search/go1.html":                  http.MethodPost,
		"/search/index.html":                http.MethodPost,
		"/search/contact.html":              http.MethodPost,
		"/search/ducks":                     http.MethodPost,
		"/search/gophers/go":                http.MethodPost,
		"/search/nature/go1.html":           http.MethodPost,
		"/search/generics/types/index.html": http.MethodPost,

		"/src/paris/invalid": http.MethodPut,
		"/src/invalid":       http.MethodPut,
		"/src1/oslo":         http.MethodPut,
		"/src1":              http.MethodPut,

		"/signal-r/protos/reflection": http.MethodPatch,
		"/signal-r":                   http.MethodPatch,
		"/signal-r/push":              http.MethodPatch,
		"/signal-r/connect":           http.MethodPatch,

		"/query/unknown/pages":           http.MethodHead,
		"/query/10/amazing/reset/single": http.MethodHead,
		"/query/911":                     http.MethodHead,
		"/query/99/sup/update-ttl":       http.MethodHead,
		"/query/46/hello":                http.MethodHead,
		"/query/unknown":                 http.MethodHead,
		"/query/untold":                  http.MethodHead,

		"/questions/1001": http.MethodConnect,
		"/questions":      http.MethodConnect,

		"/graphql":     http.MethodDelete,
		"/graph":       http.MethodDelete,
		"/graphql/cmd": http.MethodDelete,

		"/file":        http.MethodDelete,
		"/file/remove": http.MethodDelete,

		"/hero-goku": http.MethodGet,
		"/hero-thor": http.MethodGet,
	}

	requests := make([]*http.Request, 0, len(tests))

	for path, method := range tests {
		req, _ := http.NewRequest(method, path, nil)
		requests = append(requests, req)
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

func BenchmarkRouter_ServeHTTP_CaseInsensitive_WithRedirect(b *testing.B) {
	r := New()
	r.UsePathCorrectionMatch(WithRedirect())

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

	tests := map[string]string{
		"/users/finD":              http.MethodGet,
		"/users/finD/yousuf":       http.MethodGet,
		"/users/john/deletE":       http.MethodGet,
		"/users/911/updatE":        http.MethodGet,
		"/users/groupS/120/dumP":   http.MethodGet,
		"/users/groupS/230/exporT": http.MethodGet,
		"/users/deletE":            http.MethodGet,
		"/users/alL/dumP":          http.MethodGet,
		"/users/alL/exporT":        http.MethodGet,
		"/users/AnY":               http.MethodGet,

		"/seArcH":                           http.MethodPost,
		"/sEarCh/gO":                        http.MethodPost,
		"/SeArcH/Go1.hTMl":                  http.MethodPost,
		"/sEaRch/inDEx.hTMl":                http.MethodPost,
		"/SEARCH/contact.html":              http.MethodPost,
		"/SeArCh/ducks":                     http.MethodPost,
		"/sEArCH/gophers/Go":                http.MethodPost,
		"/sEArCH/nature/go1.HTML":           http.MethodPost,
		"/search/generics/types/index.html": http.MethodPost,

		"/Src/paris/InValiD": http.MethodPut,
		"/SrC/InvaliD":       http.MethodPut,
		"/SrC1/oslo":         http.MethodPut,
		"/SrC1":              http.MethodPut,

		"/Signal-R/protos/reflection": http.MethodPatch,
		"/sIgNaL-r":                   http.MethodPatch,
		"/SIGNAL-R/push":              http.MethodPatch,
		"/sIGNal-r/connect":           http.MethodPatch,

		"/quERy/unKNown/paGEs":           http.MethodHead,
		"/QUery/10/amazing/reset/SiNglE": http.MethodHead,
		"/QueRy/911":                     http.MethodHead,
		"/qUERy/99/sup/update-ttl":       http.MethodHead,
		"/QueRy/46/hello":                http.MethodHead,
		"/qUeRy/uNkNoWn":                 http.MethodHead,
		"/QuerY/UntOld":                  http.MethodHead,

		"/qUestions/1001": http.MethodOptions,
		"/quEsTioNs":      http.MethodOptions,

		"/GRAPHQL":     http.MethodDelete,
		"/gRapH":       http.MethodDelete,
		"/grAphQl/cMd": http.MethodDelete,

		"/File":        http.MethodDelete,
		"/fIle/rEmOve": http.MethodDelete,

		"/heRO-goku": http.MethodGet,
		"/HEro-thor": http.MethodGet,
	}

	requests := make([]*http.Request, 0, len(tests))

	for path, method := range tests {
		req, _ := http.NewRequest(method, path, nil)
		requests = append(requests, req)
	}

	srv := r.Serve()
	rw := newMockRW()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, request := range requests {
			srv.ServeHTTP(rw, request)
		}
	}
}

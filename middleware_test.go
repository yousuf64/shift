package shift

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteContextMiddleware(t *testing.T) {
	r := New()
	r.Use(RouteContext())
	r.GET("/foo/:name", func(w http.ResponseWriter, r *http.Request, _ Route) error {
		route := RouteOf(r)
		assert(t, route.Path == "/foo/:name", fmt.Sprintf("path > expected: /foo/:name, got: %s", route.Path))
		name := route.Params.Get("name")
		assert(t, name == "bar", fmt.Sprintf("param > expected: bar, got: %s", name))
		return nil
	})

	srv := r.Serve()
	rw := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/foo/bar", nil)
	srv.ServeHTTP(rw, req)

	assert(t, rw.Code == http.StatusOK, fmt.Sprintf("http status code > expected: 200, got: %d", rw.Code))
}

func BenchmarkRouteContextMiddleware(b *testing.B) {
	r := New()
	r.Use(RouteContext())
	r.GET("/movies/genres/:name", HTTPHandlerFunc(fakeHttpHandler))
	srv := r.Serve()

	rr := httptest.NewRecorder()
	requests := make([]*http.Request, 0, 10)
	for _, genre := range []string{"drama", "western", "sci-fi", "thriller", "animation", "adventure", "noir", "fantasy", "crime", "comedy"} {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/movies/genres/%s", genre), nil)
		requests = append(requests, req)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, req := range requests {
			srv.ServeHTTP(rr, req)
		}
	}
}

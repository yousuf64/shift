package dune

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
		assert(t, route.Template == "/foo/:name", fmt.Sprintf("template > expected: /foo/:name, got: %s", route.Template))
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

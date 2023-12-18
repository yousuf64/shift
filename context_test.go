package shift

import (
	"fmt"
	"net/http"
	"testing"
)

func TestContext_FromContext(t *testing.T) {
	t.Run("param route", func(t *testing.T) {
		ip := newInternalParams(1)
		ip.setKeys(&[]string{"name"})
		ip.appendValue("dino")

		paramRoute := Route{
			Params: newParams(ip),
			Path:   "/foo/:name",
		}
		req, _ := http.NewRequest(http.MethodGet, "/foo/dino", nil)
		req = req.WithContext(WithRoute(req.Context(), paramRoute))
		route, ok := FromContext(req.Context())

		assert(t, ok, "expected to find a route context")
		assert(t, route == paramRoute, "expected routes to be equal")
	})

	t.Run("static route", func(t *testing.T) {
		staticRoute := Route{
			Params: Params{},
			Path:   "/foo",
		}
		req, _ := http.NewRequest(http.MethodGet, "/foo", nil)
		req = req.WithContext(WithRoute(req.Context(), staticRoute))
		route, ok := FromContext(req.Context())

		assert(t, ok, "expected to find a route context")
		assert(t, route == staticRoute, "expected routes to be equal")
	})
}

func BenchmarkContext_FromContext(b *testing.B) {
	b.Run("routeCtx context", func(b *testing.B) {
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

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			FromContext(req.Context())
		}
	})

	b.Run("non-routeCtx context", func(b *testing.B) {
		req, _ := http.NewRequest(http.MethodGet, "/movies/genres/noir", nil)

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			FromContext(req.Context())
		}
	})
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

func TestContext_RouteOf(t *testing.T) {
	abcRoute := Route{
		Params: newParams(nil),
		Path:   "/abc",
	}
	req, _ := http.NewRequest("GET", "/abc", nil)
	ctx := WithRoute(req.Context(), abcRoute)
	req = req.WithContext(ctx)

	assert(t, RouteOf(req) == abcRoute, "expected routes to be equal")
}

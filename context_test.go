package shift

import (
	"context"
	"net/http"
	"testing"
)

func TestContext_FromContext(t *testing.T) {
	fooRoute := Route{
		Params: Params{emptyParams},
		Path:   "/foo",
	}
	ctx := WithRoute(context.Background(), fooRoute)
	route, ok := FromContext(ctx)

	assert(t, ok, "expected to find a route context")
	assert(t, route == fooRoute, "expected routes to have identical properties")
}

func TestContext_RouteOf(t *testing.T) {
	abcRoute := Route{
		Params: Params{nil},
		Path:   "/abc",
	}
	request, _ := http.NewRequest("GET", "/abc", nil)
	ctx := WithRoute(request.Context(), abcRoute)
	request = request.WithContext(ctx)

	assert(t, RouteOf(request) == abcRoute, "expected routes to have identical properties")
}

package shift

import (
	"context"
	"net/http"
	"sync"
)

var ctxKey uint8

// routeCtx embeds and implements context.Context.
// It is used to wrap Route object within a context.Context interface.
//
// When pooling routeCtx in a sync.Pool, make sure to reset the object before putting back to the pool.
//
//	pool := sync.Pool{...}
//	ctx = shift.WithRoute(ctx, route)
//	ctx.reset()
//	pool.Put(ctx)
type routeCtx struct {
	context.Context
	Route
}

func (ctx *routeCtx) Value(key any) any {
	if key == &ctxKey {
		return ctx
	}
	return ctx.Context.Value(key)
}

// reset resets the routeCtx values to zero values.
func (ctx *routeCtx) reset() {
	ctx.Context = nil
	ctx.Route.Params.internal = nil
	ctx.Route.Path = ""
}

// emptyRoute is a Route object with emptyParams and empty Path value.
var emptyRoute = Route{
	Params: Params{emptyParams},
	Path:   "",
}

// WithRoute returns a context.Context wrapping the provided context.Context and the Route.
func WithRoute(ctx context.Context, route Route) context.Context {
	return &routeCtx{ctx, route}
}

// FromContext unpacks Route from the provided context.Context.
// Returns false as the second return value if a Route was not found within the provided context.Context.
// Returned Route.Params can never be <nil> even if a Route is not found as it replaces <nil> Route.Params object
// with an empty internalParams object.
func FromContext(ctx context.Context) (Route, bool) {
	if rctx, ok := ctx.Value(&ctxKey).(*routeCtx); ok {
		return rctx.Route, true
	}
	return emptyRoute, false
}

// RouteOf unpacks Route information from the provided http.Request context.
// Returns an empty route if a Route was not found within the provided http.Request context.
// Returned Route.Params is always non-nil, so it's not necessary to perform a <nil> check.
// Use RouteContext middleware in the middleware stack to pack Route information into http.Request context.
//
// It is a shorthand for,
//
//	route, _ := FromContext(r.Context())
func RouteOf(r *http.Request) Route {
	route, _ := FromContext(r.Context())
	return route
}

// ctxPool pools routeCtx objects for reuse.
var ctxPool = sync.Pool{
	New: func() any {
		return &routeCtx{}
	},
}

func getCtx() *routeCtx {
	return ctxPool.Get().(*routeCtx)
}

func releaseCtx(ctx *routeCtx) {
	ctx.reset()
	ctxPool.Put(ctx)
}

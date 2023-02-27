package ape

import (
	"context"
	"net/http"
	"sync"
)

var ctxKey uint8

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

func (ctx *routeCtx) reset() {
	ctx.Context = nil
	ctx.Route.Params = nil
	ctx.Route.Template = ""
}

var emptyRoute = Route{
	Params:   emptyParams,
	Template: "",
}

// RouteOf unpacks Route information from the http.Request context.
// Use RouteContext middleware in the middleware stack to pack Route information into http.Request context.
func RouteOf(r *http.Request) Route {
	if c, ok := r.Context().Value(&ctxKey).(*routeCtx); ok {
		return c.Route
	}
	return emptyRoute
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

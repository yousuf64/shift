package dune

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
	if key == ctxKey {
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

// RouteOf unpacks Route information from the http.Request's Context.
//
// Use RouteContextMiddleware in the middleware chain to pack Route information into http.Request's Context.
func RouteOf(r *http.Request) Route {
	if c, ok := r.Context().Value(ctxKey).(*routeCtx); ok {
		return c.Route
	}
	return emptyRoute
}

var ctxPool = sync.Pool{
	New: func() any {
		return &routeCtx{}
	},
}

func getCtxFromPool() *routeCtx {
	return ctxPool.Get().(*routeCtx)
}

func putCtxToPool(ctx *routeCtx) {
	ctx.reset()
	ctxPool.Put(ctx)
}

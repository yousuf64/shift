package ape

type core = Core

type Group struct {
	core
}

// Use attaches middlewares to the current middleware stack.
// The middleware stack is executed before the request handler in the order middlewares were registered.
//
// Make sure to register middlewares before registering routes. Otherwise, the routes registered prior to registering
// middlewares wouldn't be executing the middlewares.
//
// Alternatively, Router.With() can be used to register middlewares for a whole group or a specific route.
//
// To use a net/http idiomatic middleware, wrap the middleware in the MiddlewareAdapter.
func (g *Group) Use(middlewares ...MiddlewareFunc) {
	g.mws = append(g.mws, middlewares...)
}

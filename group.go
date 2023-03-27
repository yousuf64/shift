package shift

import "strings"

type core = Core

// Group builds on top of Core and provides additional Group specific methods.
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
// To use a net/http idiomatic middleware, wrap the middleware in the HTTPMiddlewareFunc.
func (g *Group) Use(middlewares ...MiddlewareFunc) {
	g.mws = append(g.mws, middlewares...)
}

// Base returns the base path of the Group.
//
// For example,
//
//	router.Group("/v1/foo", func(group *shift.Group) {
//		group.Base() // returns /v1/foo
//	})
func (g *Group) Base() string {
	return g.base
}

// Routes returns the routes registered within the Group.
// To retrieve all the routes, call Routes() from the Router.
func (g *Group) Routes() (routes []RouteInfo) {
	for _, log := range *g.logs {
		if strings.HasPrefix(log.path, g.base) {
			routes = append(routes, RouteInfo{
				Method: log.method,
				Path:   log.path,
			})
		}
	}
	return
}

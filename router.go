package dune

import (
	"net/http"
	"strings"
)

type RoutingBehavior uint8

const (
	BehaviorSkip     RoutingBehavior = iota // BehaviorSkip action opts-out of the feature.
	BehaviorRedirect                        // BehaviorRedirect action replies with the status 301 (http.StatusMovedPermanently) and the redirect url.
	BehaviorExecute                         // BehaviorExecute action executes the matched route handler immediately.
)

func validateBehavior(behavior RoutingBehavior) {
	if behavior > 2 {
		panic("invalid routing behavior")
	}
}

type Config struct {
	trailingSlashMatch     RoutingBehavior
	sanitizeUrlMatch       RoutingBehavior
	notFoundHandler        func(w http.ResponseWriter, r *http.Request)
	handleMethodNotAllowed bool
}

var defaultConfig = &Config{
	trailingSlashMatch:     BehaviorSkip,
	sanitizeUrlMatch:       BehaviorSkip,
	notFoundHandler:        http.NotFound,
	handleMethodNotAllowed: false,
}

type group = Group

type Router struct {
	group

	config *Config
}

func New() *Router {
	d :=
		&Router{
			Group{
				Core{
					"",
					&[]routeLog{},
					nil,
				},
			},
			&Config{
				defaultConfig.trailingSlashMatch,
				defaultConfig.sanitizeUrlMatch,
				defaultConfig.notFoundHandler,
				defaultConfig.handleMethodNotAllowed,
			},
		}

	return d
}

// UseTrailingSlashMatch enables searching for a handler with/without the trailing slash when a match has not been found for the current route.
// The execution behavior is determined by the set RoutingBehavior.
//
// The default behavior is BehaviorSkip, which opt-outs of this feature.
func (r *Router) UseTrailingSlashMatch(behavior RoutingBehavior) {
	validateBehavior(behavior)
	r.config.trailingSlashMatch = behavior
}

// UseSanitizeURLMatch enables searching for a handler with the URL sanitized when a match has not been found for the current route.
// The execution behavior is determined by the set RoutingBehavior.
//
// The default behavior is BehaviorSkip, which opt-outs of this feature.
func (r *Router) UseSanitizeURLMatch(behavior RoutingBehavior) {
	validateBehavior(behavior)
	r.config.sanitizeUrlMatch = behavior
}

// UseMethodNotAllowedHandler responds with HTTP status 405 and a list of registered HTTP methods for the path in the 'Allow' header
// when a match has not been found but the path has been registered for other HTTP methods.
func (r *Router) UseMethodNotAllowedHandler() {
	r.config.handleMethodNotAllowed = true
}

// UseNotFoundHandler registers the handler to execute when a route match is not found.
func (r *Router) UseNotFoundHandler(f func(w http.ResponseWriter, r *http.Request)) {
	r.config.notFoundHandler = f
}

type RouteInfo struct {
	Method string
	Path   string
}

// Routes returns all the registered routes.
// To retrieve only the routes registered within the current scope, use RoutesScoped instead.
func (r *Router) Routes() (routes []RouteInfo) {
	routes = make([]RouteInfo, 0, len(*r.logs))

	for _, log := range *r.logs {
		routes = append(routes, RouteInfo{
			Method: log.method,
			Path:   log.path,
		})
	}

	return
}

// RoutesScoped returns the routes registered within the current scope.
// To retrieve all the routes, use Routes instead.
func (r *Router) RoutesScoped() (routes []RouteInfo) {
	routes = make([]RouteInfo, 0, len(*r.logs))
	all := len(r.base) == 0

	for _, log := range *r.logs {
		if all || strings.HasPrefix(log.path, r.base) {
			last := len(routes)
			routes = routes[:last+1]
			routes[last] = RouteInfo{
				Method: log.method,
				Path:   log.path,
			}
		}
	}

	return
}

// Base returns the base path of the group.
//
// For example,
//
//	d.Group("/v1/foo", func(d *dune.Dune) {
//		d.Base() # returns /v1/foo
//	})
func (r *Router) Base() string {
	return r.base
}

type methodInfo struct {
	staticRoutes int
	logs         []routeInfo
}

type routeInfo struct {
	method  string
	path    string
	handler HandlerFunc
	static  bool
}

var builtInMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
	http.MethodTrace,
	http.MethodConnect,
}

// Serve generates the Server which implements http.Handler interface.
func (r *Router) Serve() *Server {
	svr := &Server{
		[9]muxInterface{},
		nil,
		nil,
		r.config,
	}

	methods := groupLogsByMethods(*r.logs)
	svr.populateRoutes(methods)

	return svr
}

func groupLogsByMethods(logs []routeLog) (methodInfoMap map[string]*methodInfo) {
	methodInfoMap = map[string]*methodInfo{}
	var anyRoutes []routeLog

	for _, log := range logs {
		if log.method == "" {
			anyRoutes = append(anyRoutes, log)
			continue
		}

		info, ok := methodInfoMap[log.method]
		if !ok {
			info = &methodInfo{
				staticRoutes: 0,
				logs:         nil,
			}
			methodInfoMap[log.method] = info
		}

		static := isStatic(log.path)
		if static {
			info.staticRoutes++
		}

		info.logs = append(info.logs, routeInfo{
			method:  log.method,
			path:    log.path,
			handler: log.handler,
			static:  static,
		})
	}

	if len(anyRoutes) > 0 {
		// Populate with all the built-in methods.
		for _, method := range builtInMethods {
			if _, ok := methodInfoMap[method]; !ok {
				methodInfoMap[method] = &methodInfo{
					staticRoutes: 0,
					logs:         nil,
				}
			}
		}

		for _, route := range anyRoutes {
			static := isStatic(route.path)

			for method, info := range methodInfoMap {
				info.logs = append(info.logs, routeInfo{
					method:  method,
					path:    route.path,
					handler: route.handler,
					static:  static,
				})

				if static {
					info.staticRoutes++
				}
			}
		}
	}

	return
}

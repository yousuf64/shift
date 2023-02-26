package dune

import (
	"net/http"
	"strings"
)

type routingBehavior uint8

const (
	behaviorSkip routingBehavior = iota
	behaviorRedirect
	behaviorExecute
)

type Config struct {
	trailingSlashMatch     *actionConfig
	sanitizeUrlMatch       *actionConfig
	notFoundHandler        func(w http.ResponseWriter, r *http.Request)
	handleMethodNotAllowed bool
}

var defaultConfig = &Config{
	trailingSlashMatch: &actionConfig{
		behavior: behaviorSkip,
		code:     0,
	},
	sanitizeUrlMatch: &actionConfig{
		behavior: behaviorSkip,
		code:     0,
	},
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
// Use WithExecute, WithRedirect or WithRedirectCustom to set the behavior.
func (r *Router) UseTrailingSlashMatch(opt ActionOption) {
	opt.apply(r.config.trailingSlashMatch)
}

// UseSanitizeURLMatch enables searching for a handler after sanitized the URL when a match has not been found for the current route.
// Use WithExecute, WithRedirect or WithRedirectCustom to set the behavior.
func (r *Router) UseSanitizeURLMatch(opt ActionOption) {
	opt.apply(r.config.sanitizeUrlMatch)
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

type actionConfig struct {
	behavior routingBehavior
	code     int
}

type ActionOption interface {
	apply(c *actionConfig)
}

type actionOption func(c *actionConfig)

func (o actionOption) apply(c *actionConfig) {
	o(c)
}

// WithExecute executes the matched request handler immediately.
func WithExecute() ActionOption {
	return actionOption(func(c *actionConfig) {
		c.behavior = behaviorExecute
		c.code = 0
	})
}

// WithRedirect writes the status code 301 (http.StatusMovedPermanently) and the redirect url to the header.
func WithRedirect() ActionOption {
	return WithRedirectCustom(http.StatusMovedPermanently)
}

// WithRedirectCustom writes the provided status code and the redirect url to the header.
// statusCode should be in the range 3XX.
func WithRedirectCustom(statusCode int) ActionOption {
	if statusCode < 300 || statusCode > 399 {
		panic("status code should be in the range 3XX")
	}
	return actionOption(func(c *actionConfig) {
		c.behavior = behaviorRedirect
		c.code = statusCode
	})
}

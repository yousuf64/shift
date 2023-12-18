package shift

import "net/http"

// Route provides route information.
// Route.Params is always non-nil, so it's not necessary to perform a <nil> check.
//
// When passing Route to a goroutine, make to sure pass a copy (use Copy method)
// instead of the original Route object. The reason being Route.Params is pooled into a sync.Pool when the
// request is completed.
type Route struct {
	Params Params
	Path   string
}

// Copy returns a copy of the [Route].
// It calls [Params.Copy] implicitly to copy the underlying [Route.Params] object.
func (r Route) Copy() Route {
	return Route{
		Params: r.Params.Copy(),
		Path:   r.Path,
	}
}

// HandlerFunc is an extension of the http.HandlerFunc signature taking a third parameter to provide route information
// and returns an error to ease global error handling in the middleware stack.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, route Route) error

// MiddlewareFunc takes a HandlerFunc and returns a HandlerFunc which can call the provided HandlerFunc.
// This design is useful for chaining handlers and building the middleware stack.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// HTTPHandlerFunc allows to use an idiomatic http.HandlerFunc in place of a HandlerFunc.
// To retrieve Route information,
//
// 1. Use RouteContext middleware to pack Route information into the http.Request context.
//
//	router.Use(shift.RouteContext())
//
// 2. Use RouteOf func to unpack from the http.Request context.
//
//	func(w http.ResponseWriter, r *http.Request) error {
//		route := RouteOf(r)
//		...
//	}
func HTTPHandlerFunc(handler http.HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		handler.ServeHTTP(w, r)
		return nil
	}
}

// HTTPMiddlewareFunc allows to use an idiomatic middleware function
// 'func(next http.Handler) http.Handler' in place of a MiddlewareFunc.
// To retrieve Route information, use the RouteContext middleware and RouteOf func.
func HTTPMiddlewareFunc(mw func(next http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) (err error) {
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = next(w, r, route)
			})).ServeHTTP(w, r)
			return
		}
	}
}

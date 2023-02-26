package dune

import "net/http"

type Route struct {
	Params   *Params
	Template string
}

func (r Route) Copy() Route {
	return Route{
		Params:   r.Params.Copy(),
		Template: r.Template,
	}
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, route Route) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// HandlerAdapter allows to use an idiomatic http.HandlerFunc in place of a HandlerFunc.
// To retrieve Route information, use the RouteContext middleware and RouteOf func.
func HandlerAdapter(handler http.HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		handler.ServeHTTP(w, r)
		return nil
	}
}

// MiddlewareAdapter allows to use an idiomatic middleware function
// 'func(next http.Handler) http.Handler' in place of a MiddlewareFunc.
// To retrieve Route information, use the RouteContext middleware and RouteOf func.
func MiddlewareAdapter(mw func(next http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) (err error) {
			mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				err = next(w, r, route)
			})).ServeHTTP(w, r)
			return
		}
	}
}

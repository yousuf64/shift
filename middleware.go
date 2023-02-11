package dune

import "net/http"

func RecoverMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		defer func() {
			recover()
		}()
		return next(w, r, route)
	}
}

// RouteContextMiddleware packs Route information into http.Request's Context.
//
// Use RouteOf to unpack Route information.
//
// It is highly recommended to use this middleware before the RecoverMiddleware.
func RouteContextMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) (err error) {
		ctx := getCtxFromPool()
		ctx.Context = r.Context()
		ctx.Route = route

		r = r.WithContext(ctx)
		err = next(w, r, route)

		putCtxToPool(ctx)
		return
	}
}

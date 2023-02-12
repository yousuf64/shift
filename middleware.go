package dune

import "net/http"

func Recover(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) error {
		defer func() {
			recover()
		}()
		return next(w, r, route)
	}
}

// RouteContext packs Route information into http.Request context.
//
// Use RouteOf to unpack Route information from the http.Request context.
//
// It is highly recommended to use this middleware before the Recover middleware to lower memory footprints in case of a panic.
func RouteContext(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route Route) (err error) {
		ctx := getCtx()
		ctx.Context = r.Context()
		ctx.Route = route

		r = r.WithContext(ctx)
		err = next(w, r, route)

		releaseCtx(ctx)
		return
	}
}

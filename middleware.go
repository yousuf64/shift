package dune

import (
	"net/http"
	"runtime/debug"
)

// Recover gracefully handle panics in the subsequent middlewares in the chain and the request handler.
// It returns HTTP 500 (Internal Server Error) status and the stack trace.
func Recover() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, route Route) (err error) {
			defer func() {
				rec := recover()
				switch rec {
				case nil:
					// do nothing.
				case http.ErrAbortHandler:
					panic(rec)
				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write(debug.Stack())
				}
			}()

			return next(w, r, route)
		}
	}
}

// RouteContext packs Route information into http.Request context.
//
// Use RouteOf to unpack Route information from the http.Request context.
//
// It is highly recommended to use this middleware before the Recover middleware to lower memory footprints in case of a panic.
func RouteContext() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
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
}

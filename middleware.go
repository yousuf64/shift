package shift

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

// Recover gracefully handle panics in the subsequent middlewares in the chain and the request handler.
// It returns HTTP 500 ([http.StatusInternalServerError]) status and write the stack trace to [os.Stderr].
//
// Use [RecoverWithWriter] to write to a different [io.Writer].
func Recover() MiddlewareFunc {
	return RecoverWithWriter(os.Stderr)
}

// RecoverWithWriter gracefully handle panics in the subsequent middlewares in the chain and the request handler.
// It returns HTTP 500 ([http.StatusInternalServerError]) status and write the stack trace to the provided [io.Writer].
//
// Use [Recover] to write to [os.Stderr].
func RecoverWithWriter(w io.Writer) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request, route Route) error {
			defer func() {
				rec := recover()
				switch rec {
				case nil:
					// do nothing.
				case http.ErrAbortHandler:
					panic(rec)
				default:
					writeStack(w, rec, 3)
					rw.WriteHeader(http.StatusInternalServerError)
				}
			}()

			return next(rw, r, route)
		}
	}
}

func writeStack(w io.Writer, rec any, skipFrames int) {
	buf := &bytes.Buffer{}

	for i := skipFrames; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		fmt.Fprintf(buf, "%s\n", f.Name())
		fmt.Fprintf(buf, "	%s:%d (%#x)\n", file, line, pc)
	}

	fmt.Fprintf(w, "panic: %v\n%s", rec, buf.String())
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

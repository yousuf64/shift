package shift

import (
	"net/http"
	"sort"
	"strings"
)

type Server struct {
	muxes       [9]multiplexer         // Muxes for default HTTP methods.
	muxIndices  []int                  // Indices of non-nil muxes. This index is useful to skip <nil> muxes.
	customMuxes map[string]multiplexer // Muxes for custom HTTP methods.
	config      *Config
}

func (svr *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}

	var mux multiplexer
	if idx := methodIndex(r.Method); idx >= 0 {
		mux = svr.muxes[idx]
	} else {
		mux = svr.customMuxes[r.Method]
	}

	if mux == nil {
		if svr.config.handleMethodNotAllowed {
			svr.handleMethodNotAllowed(path, r.Method, w)
		}

		svr.config.notFoundHandler(w, r)
		return
	}

	handler, ps, template := mux.find(path)
	if handler != nil {
		if ps == nil {
			// Replace with immutable empty params object.
			// This is to ensure Route.Params is never <nil> in the request handler.
			ps = emptyParams
		}

		_ = handler(w, r, Route{
			Params: ps,
			Path:   template,
		})
		return
	}

	// Look with/without trailing slash.
	if svr.config.trailingSlashMatch.behavior != behaviorSkip {
		var clean string
		if len(path) > 0 && path[len(path)-1] == '/' {
			clean = path[:len(path)-1]
		} else {
			clean = path + "/"
		}

		handler, ps, template = mux.find(clean)
		if handler != nil {
			switch svr.config.trailingSlashMatch.behavior {
			case behaviorRedirect:
				r.URL.Path = clean
				http.Redirect(w, r, r.URL.String(), svr.config.trailingSlashMatch.code)
				return
			case behaviorExecute:
				if ps == nil {
					// Replace with immutable empty params object.
					// This is to ensure Route.Params is never <nil> in the request handler.
					ps = emptyParams
				}
				r.URL.Path = clean
				_ = handler(w, r, Route{
					Params: ps,
					Path:   template,
				})
				return
			}
		}
	}

	// Correct the path and do a case-insensitive search...
	if svr.config.pathCorrectionMatch.behavior != behaviorSkip {
		clean := cleanPath(path)
		handler, ps, matchedPath := mux.findCaseInsensitive(clean, svr.config.pathCorrectionMatch.behavior == behaviorExecute)
		if handler != nil {
			switch svr.config.pathCorrectionMatch.behavior {
			case behaviorRedirect:
				r.URL.Path = matchedPath
				http.Redirect(w, r, r.URL.String(), svr.config.pathCorrectionMatch.code)
				return
			case behaviorExecute:
				if ps == nil {
					// Replace with immutable empty params object.
					// This is to ensure Route.Params is never <nil> in the request handler.
					ps = emptyParams
				}
				r.URL.Path = matchedPath
				_ = handler(w, r, Route{
					Params: ps,
					Path:   "",
				})
				return
			}
		}
	}

	// Look for allowed methods.
	if svr.config.handleMethodNotAllowed {
		svr.handleMethodNotAllowed(path, r.Method, w)
	}

	svr.config.notFoundHandler(w, r)
	return
}

func (svr *Server) populateRoutes(byMethods map[string]*methodInfo) {
	for method, info := range byMethods {
		var mux multiplexer

		total := len(info.logs) // Total routes in the method.
		staticPercentage := float64(info.staticRoutes) / float64(total) * 100

		// Determine mux variant.
		if staticPercentage == 100 {
			mux = newStaticMux()
		} else if staticPercentage >= 30 {
			mux = newHybridMux()
		} else {
			mux = newRadixMux()
		}

		// Register routes.
		for _, log := range info.logs {
			mux.add(log.path, log.static, log.handler)
		}

		// Store mux.
		if idx := methodIndex(method); idx >= 0 {
			svr.muxes[idx] = mux

			// Store indices of active muxes in ascending order.
			svr.muxIndices = append(svr.muxIndices, idx)
			sort.Slice(svr.muxIndices, func(i, j int) bool {
				return svr.muxIndices[i] < svr.muxIndices[j]
			})
		} else {
			if svr.customMuxes == nil {
				svr.customMuxes = make(map[string]multiplexer)
			}
			svr.customMuxes[method] = mux
		}
	}

}

func (svr *Server) handleMethodNotAllowed(path string, method string, w http.ResponseWriter) {
	allowed := svr.allowedHeader(path, method)

	if len(allowed) > 0 {
		w.Header().Add("Allow", allowed)
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (svr *Server) allowedHeader(path string, skipMethod string) string {
	var allowed strings.Builder
	skipped := false

	if skipMethodIdx := methodIndex(skipMethod); skipMethodIdx != -1 {
		for _, idx := range svr.muxIndices {
			if !skipped && idx == skipMethodIdx {
				skipped = true
				continue
			}

			if handler, _, _ := svr.muxes[idx].find(path); handler != nil {
				if allowed.Len() != 0 {
					allowed.WriteString(", ")
				}

				allowed.WriteString(methodString(idx))
			}
		}
	}

	for method, mux := range svr.customMuxes {
		if !skipped && method == skipMethod {
			continue
		}

		if handler, _, _ := mux.find(path); handler != nil {
			if allowed.Len() != 0 {
				allowed.WriteString(", ")
			}

			allowed.WriteString(method)
		}
	}

	return allowed.String()
}

func methodIndex(method string) int {
	switch method {
	case http.MethodGet:
		return 0
	case http.MethodPost:
		return 1
	case http.MethodPut:
		return 2
	case http.MethodPatch:
		return 3
	case http.MethodDelete:
		return 4
	case http.MethodHead:
		return 5
	case http.MethodOptions:
		return 6
	case http.MethodTrace:
		return 7
	case http.MethodConnect:
		return 8
	default:
		return -1
	}
}

func methodString(idx int) string {
	switch idx {
	case 0:
		return http.MethodGet
	case 1:
		return http.MethodPost
	case 2:
		return http.MethodPut
	case 3:
		return http.MethodPatch
	case 4:
		return http.MethodDelete
	case 5:
		return http.MethodHead
	case 6:
		return http.MethodOptions
	case 7:
		return http.MethodTrace
	case 8:
		return http.MethodConnect
	default:
		return ""
	}
}

package dune

import "strings"

type routeScanner struct {
	path string
	low  int
	high int
	wc   bool
}

func newRouteScanner(path string) *routeScanner {
	return &routeScanner{
		path: path,
		wc:   len(path) > 0 && (path[0] == ':' || path[0] == '*'),
	}
}

func (r *routeScanner) next() string {
	if r.high > len(r.path)-1 {
		return ""
	}

Loop:
	for r.high < len(r.path) {
		if r.wc {
			if r.path[r.high] == '/' {
				break Loop
			}
		} else {
			switch r.path[r.high] {
			case ':', '*':
				break Loop
			}
		}

		r.high++
	}

	r.wc = !r.wc
	seg := r.path[r.low:r.high]

	r.low = r.high

	return seg
}

func (r *routeScanner) indexOf(c uint8) int {
	return strings.IndexByte(r.path, c)
}

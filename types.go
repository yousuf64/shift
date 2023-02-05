package dune

import "net/http"

type Route struct {
	Params   *Params
	Template string
}

type HandlerFunc func(w http.ResponseWriter, r *http.Request, route Route) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

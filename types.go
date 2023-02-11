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

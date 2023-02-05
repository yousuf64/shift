package dune

type core = Core

type Group struct {
	core
}

func (g *Group) Use(middlewares ...MiddlewareFunc) {
	g.mws = append(g.mws, middlewares...)
}

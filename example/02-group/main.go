package main

import (
	"github.com/yousuf-git/dune-project"
	"net/http"
)

func main() {
	r := dune.New()
	r.Group("/v1", func(g *dune.Group) {
		g.GET("/abc", func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
			_, err := w.Write([]byte("inline group v1: abc"))
			return err
		})
		g.GET("/xyz", func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
			_, err := w.Write([]byte("inline group v1: xyz"))
			return err
		})
	})
	r.Group("/v2", groupV2)
	r.Group("/v3", groupV3)
	r.Group("/v4", func(g *dune.Group) {
		g.Group("/aaa", func(g *dune.Group) {
			g.Group("/bbb", func(g *dune.Group) {
				g.Group("/ccc", func(g *dune.Group) {
					g.GET("", func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
						_, err := w.Write([]byte("response from nested group"))
						return err
					})
				})
			})
		})
	})

	_ = http.ListenAndServe(":6464", r.Serve())
}

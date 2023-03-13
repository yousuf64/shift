package main

import (
	"github.com/yousuf64/shift"
	"net/http"
)

func main() {
	r := shift.New()
	r.Group("/v1", func(g *shift.Group) {
		g.GET("/abc", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
			_, err := w.Write([]byte("inline group v1: abc"))
			return err
		})
		g.GET("/xyz", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
			_, err := w.Write([]byte("inline group v1: xyz"))
			return err
		})
	})
	r.Group("/v2", groupV2)
	r.Group("/v3", groupV3)
	r.Group("/v4", func(g *shift.Group) {
		g.Group("/aaa", func(g *shift.Group) {
			g.Group("/bbb", func(g *shift.Group) {
				g.Group("/ccc", func(g *shift.Group) {
					g.GET("", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
						_, err := w.Write([]byte("response from nested group"))
						return err
					})
				})
			})
		})
	})

	_ = http.ListenAndServe(":6464", r.Serve())
}

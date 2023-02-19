package main

import (
	"github.com/yousuf-git/dune-project"
	"net/http"
)

func groupV2(g *dune.Group) {
	g.GET("/abc", func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
		_, err := w.Write([]byte("v2.go file: abc"))
		return err
	})
	g.GET("/xyz", func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
		_, err := w.Write([]byte("v2.go file: xyz"))
		return err
	})
}
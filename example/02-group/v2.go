package main

import (
	"github.com/yousuf64/ape"
	"net/http"
)

func groupV2(g *ape.Group) {
	g.GET("/abc", func(w http.ResponseWriter, r *http.Request, route ape.Route) error {
		_, err := w.Write([]byte("v2.go file: abc"))
		return err
	})
	g.GET("/xyz", func(w http.ResponseWriter, r *http.Request, route ape.Route) error {
		_, err := w.Write([]byte("v2.go file: xyz"))
		return err
	})
}

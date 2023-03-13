package main

import (
	"fmt"
	"github.com/yousuf64/shift"
	"net/http"
)

func main() {
	r := shift.New()
	// Static route.
	r.GET("/", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte("hello from shift"))
		return err
	})
	// Param route.
	r.GET("/user/:name", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte(fmt.Sprintf("hello %s", route.Params.Get("name"))))
		return err
	})
	// Mid-segment param route.
	r.DELETE("/version:number", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte(fmt.Sprintf("version %s deleted", route.Params.Get("number"))))
		return err
	})
	// Wildcard route.
	r.HEAD("/bucket/*path", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte(fmt.Sprintf("file found at %s", route.Params.Get("path"))))
		return err
	})
	// Mid-segment wildcard route.
	r.PUT("/vid*url", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte(fmt.Sprintf("fetched video from %s", route.Params.Get("url"))))
		return err
	})
}

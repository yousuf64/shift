package main

import (
	"errors"
	"fmt"
	"github.com/yousuf64/shift"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	r := shift.New()
	r.Use(traceMiddleware) // Apply to all the handlers declared after Use().

	r.GET("/", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte("hello from shift"))
		return err
	})

	// Apply only to the subsequently chained handler.
	r.With(timezoneMiddleware).GET("/bar", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		_, err := w.Write([]byte(fmt.Sprintf("client timezone: %s", r.Header.Get("Timezone"))))
		return err
	})

	// Apply only to the subsequently chained group.
	r.With(rateLimiterMiddleware).Group("/foo", func(g *shift.Group) {
		g.GET("/aaa", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
			_, err := w.Write([]byte(":)"))
			return err
		})
	})

	r.Group("/oof", func(g *shift.Group) {
		g.Use(authMiddleware) // Apply to all the handlers declared after Use() within this group scope.
		g.GET("/aaa", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
			_, err := w.Write([]byte("hello from authenticated route"))
			return err
		})
	})

	_ = http.ListenAndServe(":6464", r.Serve())
}

var i int64 = 0 // Fake Trace ID

func traceMiddleware(next shift.HandlerFunc) shift.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route shift.Route) (err error) {
		id := i
		atomic.AddInt64(&i, 1)
		w.Header().Set("Trace-ID", strconv.Itoa(int(id)))

		u := r.URL.String()
		t := time.Now()
		log.Printf("received request | id: %d, url: %s", id, u)

		err = next(w, r, route)

		log.Printf("completed request | id: %d, url: %s, time elapsed: %dμs", id, u, time.Since(t).Microseconds())
		return
	}
}

func timezoneMiddleware(next shift.HandlerFunc) shift.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		if v := r.Header.Get("Timezone"); v != "" {
			return next(w, r, route)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("'Timezone' header required"))
		return errors.New("'Timezone' header required")
	}
}

var count int64 = 0
var threshold = 5
var interval = time.Second * 10
var timer = time.NewTimer(interval)

func init() {
	go func() {
		for range timer.C {
			count = 0
			timer.Reset(interval)
		}
	}()
}

// accepts only x number of requests within y time period.
func rateLimiterMiddleware(next shift.HandlerFunc) shift.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		if count >= int64(threshold) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(fmt.Sprintf("try again in few seconds")))
			return nil
		}
		atomic.AddInt64(&count, 1)
		return next(w, r, route)
	}
}

func authMiddleware(next shift.HandlerFunc) shift.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
		if v := r.Header.Get("Authorization"); v != "" {
			return next(w, r, route)
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("'Authorization' header required"))
		return errors.New("'Authorization' header required")
	}
}

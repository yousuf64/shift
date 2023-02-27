package main

import (
	"errors"
	"github.com/yousuf64/ape"
	"net/http"
)

func main() {
	r := ape.New()
	r.Use(errorHandler)
	r.GET("/order", func(w http.ResponseWriter, r *http.Request, route ape.Route) error {
		return errors.New("unable to publish the event")
	})
	r.GET("/pay", func(w http.ResponseWriter, r *http.Request, route ape.Route) error {
		return customError{
			StatusCode: http.StatusPaymentRequired,
			Message:    "missing payment method",
		}
	})

	_ = http.ListenAndServe(":6464", r.Serve())
}

type customError struct {
	StatusCode int
	Message    string
}

func (e customError) Error() string {
	return e.Message
}

func errorHandler(next ape.HandlerFunc) ape.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, route ape.Route) error {
		err := next(w, r, route)
		if err != nil {
			switch err := err.(type) {
			case customError:
				w.WriteHeader(err.StatusCode)
				_, _ = w.Write([]byte(err.Message))
			default:
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
			}
		}

		return nil
	}
}

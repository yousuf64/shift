# dune 🚀

`dune` is a lightweight blistering fast HTTP router for Go. It's designed with simplicity and performance in mind. It uses radix trees and hash maps with lots of indexing under the hood to achieve high performance.

## Benchmarks


## Install

```
go get -u github.com/yousuf-git/dune-project
```

## Features
* Fast and zero heap allocations
* Middleware support
* `net/http` compatible (Both request handlers and middlewares)
* Route grouping
* Powerful routing system
  * Route prioritization (Static > Param > Wildcard)
  * Case-insensitive route matching
  * Trailing slash (with/without) route matching
  * Path autocorrection
  * No route conflict limitations (`/posts/:id` and `/posts/export` is allowed)
  * Allows different param names over the same path (`/users/:name` and `/users/:id/delete` is valid)
  * Mid-segment param (`/v:version/jobs`, `/stream_*url`)
* Lightweight
* Zero external dependencies

## Quick Start
```go
package main

import (
	"fmt"
	"github.com/yousuf-git/dune-project"
	"net/http"
)

func main() { 
	// Router
	r := dune.New()

	// Middleware
	r.Use(dune.Recover())

	// Routes
	r.GET("/", greet)

	// Run
	fmt.Println(http.ListenAndServe(":6464", r.Serve()))
}

// Handler
func greet(w http.ResponseWriter, r *http.Request, route dune.Route) error {
	_, err := w.Write([]byte("hello!"))
	return err
}

```
## Routing System
`dune` routing system is very powerful and straightforward.
```
> Pattern: /foo
    /foo              match
    /                 no match
    /foo/foo          no match

> Pattern: /user/:name
    /user/saul        match
    /user/saul/foo    no match
    /user/            no match
    /user             no match
    
> Pattern: /user:name
    /usersaul         match
    /user             no match
    
> Pattern: /user:fname:lname (not allowed, allows only one param within a segment '/.../')

> Pattern: /stream/*path
    /stream/foo/bar/abc.mp4    match
    /stream/foo                match
    /stream/                   match
    /stream                    no match
    
> Pattern: /stream*path
    /streamfoo/bar/abc.mp4    match
    /streamfoo                match
    /stream                   match
    /strea                    no match
    
> Pattern: /*url*directory (not allowed, allows only one wildcard param per route)
```

## Request Handler
`dune` uses a slightly modified version of the `net/http` request handler, with an additional parameter
that provides route information. Also, the request handler returns an error. It makes it convenient to
handle errors in middleware without cluttering the handlers.
```go
func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
	_, err := w.Write([]byte("hello world!"))
	return err
}
```

You can also use `net/http` request handlers using the `HandlerAdapter`.
```go
package main

import (
	"github.com/yousuf-git/dune-project"
	"net/http"
)

func main() {
	// ...
	
	r.GET('/', dune.HandlerAdapter(hello))
	
	// ...
}

func hello(w http.ResponseWriter, r *http.Request) {
	 _, _ = w.Write([]byte("hello world!"))
}
```

To retrieve Route information from a `net/http` handler, use the `RouteContext` middleware and `RouteOf` function.
```go
r.Use(dune.RouteContext())
r.GET('/hello/:name', dune.HandlerAdapter(hello))

func hello(w http.ResponseWriter, r *http.Request) {
    route := dune.RouteOf(r)
    route.Template // /hello/:name 
    route.Params.Get('name') // saul
}
```

## Middlewares
`dune` supports both `dune` and `net/http` style middlewares. Which means you can use any stdlib compatible middlewares.

* `dune` middleware signature: `func (next dune.HandlerFunc) dune.HandlerFunc`
* `net/http` middleware signature: `func (next http.Handler) http.Handler`

Use `MiddlewareAdapter` to bind `net/http` middleware.

To attach a middleware to the current scope, use `router.Use()`,
```go
func main() {
    // ...
    r.Use(AuthMiddleware, dune.MiddlewareAdapter(AnotherMiddleware))
    r.GET('/', hello)
    r.POST('/users', createUser)
    // ...
}

func AuthMiddleware(next dune.HandlerFunc) dune.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request, route dune.Route) error {
        // auth logic...
        // you could conditionally circuit break here without calling next().
        return next(w, r, route)
    }
}

func AnotherMiddleware(next http.Handler) http.Handler 
// ...

```
To attach a middleware to a specific request handler or a group, use `router.With()`,
```go
r.With(AuthMiddleware, dune.MiddlewareAdapter(AnotherMiddleware)).GET("/", hello)
r.With(AuthMiddleware).Group("/v1", v1Group)
```

### Built-in Middlewares

| Middleware handler | Description                                             |
|--------------------|---------------------------------------------------------|
| RouteContext       | Packs route information into `http.Request` context     |
| Recover            | Gracefully handle panics                                |

### Custom Middleware
Check out middleware examples.

## Error Handling
`dune` makes it very convenient to centralize error handling without cluttering the handlers using the middlewares.

Check out error handling examples.

## Route Information
In a `dune` style request handler, you can access route information such as the route template and route params directly through the `route` argument.

In a `net/http` style request handler, you'd have to use the `RouteContext` middleware and within the request handler, use `RouteOf` to retrieve the `route` object.
## `shift`: high-performance HTTP router for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/yousuf64/shift.svg)](https://pkg.go.dev/github.com/yousuf64/shift)
[![Go Report Card](https://goreportcard.com/badge/github.com/yousuf64/shift)](https://goreportcard.com/report/github.com/yousuf64/shift)
[![codecov](https://codecov.io/gh/yousuf64/shift/branch/main/graph/badge.svg?token=NK2KPJNYVA)](https://codecov.io/gh/yousuf64/shift)

High-performance HTTP router for Go, with a focus on speed, simplicity, and ease-of-use.

```
go get -u github.com/yousuf64/shift
```

At the core of its performance, `shift` uses a powerful combination of radix trees and hash maps, setting the standard for lightning-fast routing.

Why `shift`?

* `shift` is faster than other mainstream HTTP routers. 
* Unlike other fast routers, `shift` strives to remain idiomatic and close to the standard library as much as possible.
* Its primary focus is on routing requests quickly and efficiently, without attempting to become a full-fledged framework.
* Despite its simplicity, `shift` offers powerful routing capabilities.
* `shift` is compatible with `net/http` request handlers and middlewares.

## Benchmarks
`shift` is benchmarked against Gin and Echo in the [benchmark suite](https://github.com/yousuf64/http-routing-benchmark/).

The benchmark suite is also available as a [GitHub Action](https://github.com/yousuf64/http-routing-benchmark/actions/workflows/benchmark.yaml).

### Results
Comparison between `shift`, `gin` and `echo` as of Feb 27, 2023 on Go 1.19.4 (windows/amd64)

Benchmark system specifications:
* 12th Gen Intel Core i7-1265U vPro (12 MB cache, 10 cores, up to 4.80 GHz Turbo)
* 32 GB (2x 16 GB), DDR4-3200
* Windows 10 Enterprise 22H2
* Go 1.19.4 (windows/amd64)

```
BenchmarkShift_CaseInsensitiveAll-12             1750636               635.6 ns/op             0 B/op          0 allocs/op
BenchmarkGin_CaseInsensitiveAll-12               1000000              1066 ns/op               0 B/op          0 allocs/op
BenchmarkShift_GithubAll-12                        79966             14575 ns/op               0 B/op          0 allocs/op
BenchmarkGin_GithubAll-12                          49107             25962 ns/op            9911 B/op        154 allocs/op
BenchmarkEcho_GithubAll-12                         54187             26318 ns/op               0 B/op          0 allocs/op
BenchmarkShift_GPlusAll-12                       2492064               632.7 ns/op             0 B/op          0 allocs/op
BenchmarkGin_GPlusAll-12                         1415556               837.9 ns/op             0 B/op          0 allocs/op
BenchmarkEcho_GPlusAll-12                        1000000              1154 ns/op               0 B/op          0 allocs/op
BenchmarkShift_OverlappingRoutesAll-12            923211              1174 ns/op               0 B/op          0 allocs/op
BenchmarkGin_OverlappingRoutesAll-12              352972              4029 ns/op            1953 B/op         32 allocs/op
BenchmarkEcho_OverlappingRoutesAll-12             552678              2310 ns/op               0 B/op          0 allocs/op
BenchmarkShift_ParseAll-12                       1490170               838.6 ns/op             0 B/op          0 allocs/op
BenchmarkGin_ParseAll-12                          748366              1492 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_ParseAll-12                         697556              1829 ns/op               0 B/op          0 allocs/op
BenchmarkShift_RandomAll-12                       817633              1241 ns/op               0 B/op          0 allocs/op
BenchmarkGin_RandomAll-12                         292681              4675 ns/op            2201 B/op         34 allocs/op
BenchmarkEcho_RandomAll-12                        428557              2717 ns/op               0 B/op          0 allocs/op
BenchmarkShift_StaticAll-12                       452316              2595 ns/op               0 B/op          0 allocs/op
BenchmarkGin_StaticAll-12                         128896              9701 ns/op               0 B/op          0 allocs/op
BenchmarkEcho_StaticAll-12                        106158             10877 ns/op               0 B/op          0 allocs/op
```

* Column 1: Benchmark name
* Column 2: Number of iterations, higher means more confident result
* Column 3: Nanoseconds elapsed per operation (ns/op), lower is better
* Column 4: Number of bytes allocated on heap per operation (B/op), lower is better
* Column 5: Average allocations per operation (allocs/op), lower is better

## Features
* Fast and zero heap allocations.
* Middleware support.
* Compatible with `net/http` request handlers and middlewares.
* Route grouping.
* Allows declaring custom HTTP methods.
* Powerful routing system that includes:
    * Route prioritization (Static > Param > Wildcard in that order).
    * Case-insensitive route matching.
    * Trailing slash with (or without) route matching.
    * Path autocorrection.
    * Allows conflicting/overlapping routes (`/posts/:id` and `/posts/export` can exist together).
    * Allows different param names over the same path (`/users/:name` and `/users/:id/delete` can exist without param name conflicts).
    * Mid-segment params (`/v:version/jobs`, `/stream_*url`).
* Lightweight.
* Has zero external dependencies.

## Quick Start
To install `shift`, simply run:
```
go get -u github.com/yousuf64/shift
```

Using `shift` is easy. Here's a simple example:

```go
package main

import (   
    "fmt"
    "github.com/yousuf64/shift"
    "net/http"
)

func main() {
    router := shift.New()
	
    router.GET("/", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
        _, err := fmt.Fprint(w, "Hello, world!")
        return err
    })
	
    http.ListenAndServe(":8080", router.Serve())
}
```

In this example, we create a `shift` router, define a GET route for the root path, and start an HTTP server to listen for incoming requests on port `8080`.

## Routing System
`shift` boasts a highly powerful and flexible routing system.
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
`shift` uses a slightly modified version of the `net/http` request handler, which includes an additional parameter providing route information. 
Moreover, the `shift` request handler can return an error, making it convenient to handle errors in middleware without cluttering the handlers.

```go
func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
    _, err := fmt.Fprintf(w, "Hello 👋")
    return err
}
```

You can also use `net/http` request handlers using `HTTPHandlerFunc` adapter.

```go
package main

import (
    "fmt"
    "github.com/yousuf64/shift"
    "net/http"
)

func main() {
    router := shift.New()
	
    // Wrap the net/http handler in HTTPHandlerFunc 
    router.GET("/", shift.HTTPHandlerFunc(HelloHandler))
	
    // ...
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
    _, _ = fmt.Fprintf(w, "👋👋👋")
}
```

To retrieve route information from a `net/http` request handler, follow these steps:
1. Attach the `RouteContext` middleware to the router, which will pack route information into the `http.Request` context.
2. In the request handler, use the `RouteOf()` function to retrieve the `Route` object from the `http.Request` context.

```go
router := shift.New()
router.Use(shift.RouteContext())
router.GET("/hello/:name", shift.HTTPHandlerFunc(HelloUserHandler))

func HelloUserHandler(w http.ResponseWriter, r *http.Request) {
    route := shift.RouteOf(r)
    _, _ = fmt.Fprintf(w, "Hello, %s 😎 from %s route", route.Params.Get("name"), route.Path)
    // Writes 'Hello, Max 😎 from /hello/:name route'
}
```

## Middlewares
`shift` supports both `shift`-style and `net/http`-style middlewares, allowing you to attach any `net/http` compatible middleware.
* The `shift` middleware signature is: `func(next shift.HandlerFunc) shift.HandlerFunc`
* The `net/http` middleware signature is: `func(next http.Handler) http.Handler`

Middlewares can be scoped to all routes, to a specific group, or even to a single route.

```go
func main() {
    router := shift.New()
	
    // Attaches to routes declared after Router.Use() statement. 
    router.Use(AuthMiddleware, shift.HTTPMiddlewareFunc(TraceMiddleware))
	
    router.GET("/", Hello)
    router.POST("/users", CreateUser)
	
    // Attaches to routes declared within the group. 
    router.With(LoggerMiddleware).Group("/posts", PostsGroup)
	
    // Attaches only to the chained route. 
    router.With(CORSMiddleware).GET("/comments", GetComments)
	
    // ...
}

func AuthMiddleware(next shift.HandlerFunc) shift.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
        // Authorization logic...

        // You can conditionally circuit break from a middleware by returning before calling next(). 
        if someCondition {
            return nil
        }
		
        return next(w, r, route)
    }
}

func TraceMiddleware(next http.Handler) http.Handler { ... }

func LoggerMiddleware(next shift.HandlerFunc) shift.HandlerFunc { ... }

func CORSMiddleware(next shift.HandlerFunc) shift.HandlerFunc { ... }
```

Note: 
* `Router.Use()` can also be used within a group. It will attach the provided middlewares to the routes declared within the group after the `Router.Use()` statement.
* `HTTPMiddlewareFunc` adapter can be used to attach `net/http` middleware.

### Built-in Middlewares

| Middleware handler | Description                                             |
|--------------------|---------------------------------------------------------|
| RouteContext       | Packs route information into `http.Request` context     |
| Recover            | Gracefully handle panics                                |

### Writing Custom Middleware
Check out [middleware examples](/example/03-middleware/main.go).

## Not Found Handler
By default, when a matching route is not found, it replies to the request with an HTTP 404 (Not Found) error. 

Use `Router.UseNotFoundHandler()` to register a custom not found handler.

```go
router.UseNotFoundHandler(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(410) // Replies with a 410 error.
})
```

## Method Not Allowed Handler
With this feature enabled, the router will check for matching routes for other HTTP methods when a matching route is not found.
If any are found, it replies with an HTTP 405 (Method Not Allowed) status code and includes the allowed methods in the `Allow` header.

Use `Router.UseMethodNotAllowedHandler()` to enable this feature.

```go
router := shift.New()
router.UseMethodNotAllowedHandler()

router.GET("/cake", GetCakeHandler)
router.POST("/cake", PostCakeHandler)
```

On `PUT /cake` request, since a `PUT` route is not registered for the `/cake` path,
the router will reply with an HTTP 405 (Method Not Allowed) status code and `GET, POST` in the `Allow` header.

## Error Handling
Since `shift` request handlers can return errors, it is easy to handle errors in middleware without cluttering the request handlers.
This helps to keep the request handlers clean and focused on their primary task.

Check out [error handling examples](/example/04-error-handler/main.go).

## Trailing Slash Match
When `Router.UseTrailingSlashMatch()` is set, if the router is unable to find a match for the path, it tries to find a match with or without the trailing slash.
The routing behavior for the matched route is determined by the provided `ActionOption` (See below).

When `Router.UseTrailingSlashMatch()` is set, if the router is unable to find a match for the requested path, it will try to find a match with or without the trailing slash.
The routing behavior for the matched route is determined by the provided `ActionOption`.

With `shift.WithExecute()` option, the matched fallback route handler would be executed.

```go
router := shift.New()
router.UseTrailingSlashMatch(shift.WithExecute())

router.GET("/foo", FooHandler) // Matches /foo and /foo/ 
router.GET("/bar/", BarHandler) // Matches /bar/ and /bar
```

In the above example, the first route handler matches both `/foo` and `/foo/` and the second route handler matches both `/bar/` and `/bar`.

## Path Correction & Case-Insensitive Match
When `Router.UsePathCorrectionMatch()` is set, if the router is unable to find a match for the path, it will perform path correction and case-insensitive matching in order to find a match for the requested path.
The routing behavior for the matched route is determined by the provided `ActionOption`.

With `shift.WithRedirect()` option, it will return a HTTP 304 (Moved Permanently) status with a redirect to correct URL.

```go
router := shift.New()
router.UsePathCorrectionMatch(shift.WithRedirect())

router.GET("/foo", FooHandler) // Matches /foo, /Foo, /fOO, /fOo, and so on...
router.GET("/bar/", BarHandler) // Matches /bar/, /Bar/, /bAr/, /BAR, /baR/, and so on...
```

## ActionOption

Both `UseTrailingSlashMatch` and `UsePathCorrectionMatch` expects an `ActionOption` which provides the routing behavior for the matched route, `shift` provides three behavior providers:
* `WithExecute()` - Executes the request handler of the correct route.
* `WithRedirect()` - Returns HTTP 304 (Moved Permanently) status with a redirect to correct URL in the header.
* `WithRedirectCustom(statusCode)` - Is same as `WithRedirect`, except it writes the provided status code (should be in range 3XX).

## Route Information
In a `shift` style request handler, access route information such as the route path and route params directly through the `Route` argument.

In a `net/http` style request handler, attach the `RouteContext` middleware and within the request handler, use `RouteOf()` function to retrieve the `Route` object.

### Using Route and Params in GoRoutines
When using `Route` or `Params` object in a Go Routine, make sure to get a clone using `Copy()` which is available for both the objects.

```go
func WorkerHandler(w http.ResponseWriter, r *http.Request, route shift.Route) error {
    go FooWorker(route.Copy()) // Copies the whole Route object along with the internal Params object.
    go BarWorker(route.Params.Copy()) // Copies only the Params object.
    return nil
}

func FooWorker(route shift.Route) { ... }

func BarWorker(ps *shift.Params) { ... }
```

## Registering to Multiple Methods
To register a request handler to multiple methods, use `Router.Map()`.

```go
router := shift.New()
router.Map([]string{"GET", "POST"}, "/zanzibar", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
    _, err := fmt.Fprintf(w, "👊👊👊")
    return err
})
```

This is equivalent to registering the request handler to the path `/zanzibar` by calling both `Router.GET()` and `Router.POST()`.

## Registering to a Custom HTTP Method
You can also use `Router.Map()` to register request handlers to custom HTTP methods.

```go
router := shift.New()
router.Map([]string{"FOO"}, "/products", func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
    _, err := fmt.Fprintf(w, "Hello, from %s method 👊", r.Method)
    return err
})
```

 ```shell
curl --request FOO --url '127.0.0.1:6464/products'
```

The router will reply `Hello, from FOO method 👊` for the above request.

## Credits
* Julien Schmidt for [HttpRouter](https://github.com/julienschmidt/httprouter).
  * `path.go` file is directly taken from this project for path correction.

## License
Licensed under [MIT License](/LICENSE)

Copyright (c) 2023 Mohammed Yousuf

## Status
`shift` is currently pre-1.0. Therefore, there could be minor breaking changes to stabilize the API before the initial stable release. Please open an issue if you have questions, requests, suggestions for improvements, or concerns.
It's intended to release 1.0.0 during the first week of April.
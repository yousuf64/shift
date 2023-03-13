# shift

`shift` is a lightweight blistering fast HTTP router for Go. It's designed with simplicity and performance in mind. It uses radix trees and hash maps with lots of indexing under the hood to achieve high performance.

## Benchmarks
Benchmark suite: https://github.com/yousuf64/http-routing-benchmark

Benchmark on GitHub Actions: https://github.com/yousuf64/http-routing-benchmark/actions/workflows/benchmark.yaml

Comparison between Shift, Gin and Echo as of Feb 27, 2023 on Go 1.19.4 (windows/amd64)
```
goos: windows
goarch: amd64
pkg: http-routing-benchmark
cpu: 12th Gen Intel(R) Core(TM) i7-1265U
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

## Install

```
go get -u github.com/yousuf64/shift
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
  * No route conflict/overlapping limitations (`/posts/:id` and `/posts/export` is allowed)
  * Allows different param names over the same path (`/users/:name` and `/users/:id/delete` is valid)
  * Mid-segment params (`/v:version/jobs`, `/stream_*url`)
* Lightweight
* Zero external dependencies

## Quick Start

```go
package main

import (
  "fmt"
  "github.com/yousuf64/shift"
  "net/http"
)

func main() {
  // Router
  router := shift.New()

  // Middleware
  router.Use(shift.Recover())

  // Routes
  router.GET("/", greet)

  // Run
  fmt.Println(http.ListenAndServe(":6464", router.Serve()))
}

// Handler
func greet(w http.ResponseWriter, r *http.Request, route shift.Route) error {
  _, err := w.Write([]byte("hello!"))
  return err
}

```
## Routing System
`shift` has a very powerful and flexible routing system.
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
`shift` uses a slightly modified version of the `net/http` request handler, with an additional parameter
that provides route information. Also, the request handler returns an error. It makes it convenient to
handle errors in middleware without cluttering the handlers.
```go
func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
	_, err := w.Write([]byte("hello world!"))
	return err
}
```

You can also use `net/http` request handlers using the `HTTPHandlerFunc`.

```go
package main

import (
  "github.com/yousuf64/shift"
  "net/http"
)

func main() {
  // ...

  router.GET('/', shift.HTTPHandlerFunc(hello))

  // ...
}

func hello(w http.ResponseWriter, r *http.Request) {
  _, _ = w.Write([]byte("hello world!"))
}
```

To retrieve Route information from a `net/http` handler, use the `RouteContext` middleware and `RouteOf` function.
```go
router.Use(shift.RouteContext())
router.GET('/hello/:name', shift.HTTPHandlerFunc(hello))

func hello(w http.ResponseWriter, r *http.Request) {
    route := shift.RouteOf(r)
    route.Template // /hello/:name 
    route.Params.Get('name') // saul
}
```

## Middlewares
`shift` supports both `shift` and `net/http` style middlewares. Which means you can use any stdlib compatible middlewares.

* `shift` middleware signature: `func (next shift.HandlerFunc) shift.HandlerFunc`
* `net/http` middleware signature: `func (next http.Handler) http.Handler`

Use `HTTPMiddlewareFunc` to bind `net/http` middleware.

To attach a middleware to the current scope, use `router.Use()`,
```go
func main() {
    // ...
    router.Use(AuthMiddleware, shift.HTTPMiddlewareFunc(AnotherMiddleware))
    router.GET('/', hello)
    router.POST('/users', createUser)
    // ...
}

func AuthMiddleware(next shift.HandlerFunc) shift.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request, route shift.Route) error {
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
router.With(AuthMiddleware, shift.MiddlewareAdapter(AnotherMiddleware)).GET("/", hello)
router.With(AuthMiddleware).Group("/v1", v1Group)
```

### Built-in Middlewares

| Middleware handler | Description                                             |
|--------------------|---------------------------------------------------------|
| RouteContext       | Packs route information into `http.Request` context     |
| Recover            | Gracefully handle panics                                |

### Custom Middleware
Check out [middleware examples](/example/03-middleware/main.go).

## Error Handling
`shift` makes it very convenient to centralize error handling without cluttering the handlers using the middlewares.

Check out [error handling examples](/example/04-error-handler/main.go).

## Trailing Slash, Path Autocorrection & Case-Insensitive Match
If the registered route is `/foo` and you want both `/foo` and `/foo/` to match the handler, enable the trailing slash matching feature.
```go
router := shift.New()
router.UseTrailingSlashMatch(shift.WithExecute())

router.GET('/foo', fooHandler) // Matches both /foo and /foo/
router.GET('/bar/', barHandler) // Matches both /bar/ and /bar
```

If you want `shift` to take care of sanitizing the URL path, enable URL sanitizing feature, which sanitizes the URL and perform a case-insensitive search instead of a regular search.
```go
router := shift.New()
router.UseSanitizeURLMatch(shift.WithRedirect())

router.GET('/foo', fooHandler) // Matches /foo, /Foo, /fOO, /fOo, and so on...
router.GET('/bar/', barHandler) // Matches /bar/, /Bar/, /bAr/, /BAR, /baR/, and so on...
```

Both `UseTrailingSlashMatch` and `UseSanitizeURLMatch` expects an `ActionOption` which provides the routing behavior for the fallback handler, `shift` provides three behavior providers:
* `WithExecute()` - Executes the request handler of the correct route.
* `WithRedirect()` - Return HTTP 304 (Moved Permanently) status and writes the correct path as the redirect url to the header.
* `WithRedirectCustom(statusCode)` - Is same as `WithRedirect`, except it writes the provided status code (should be in range 3XX).

## Route Information
In a `shift` style request handler, you can access route information such as the route template and route params directly through the `route` argument.

In a `net/http` style request handler, you'd have to use the `RouteContext` middleware and within the request handler, use `RouteOf` to retrieve the `route` object.

### Using Route and Params in GoRoutines
When using `Route` or `Params` object in a Go Routine, make sure to get a clone using `Copy()` which is available for both the objects.
```go
func handler(w http.ResponseWriter, r *http.Request, route shift.Route) error {
	go fooWorker(route.Copy()) // Copies the whole Route object along with the internal Params object
	go barWorker(route.Params.Copy()) // Copies only the Params object
	return nil
}

func fooWorker(route shift.Route) {}

func barWorker(ps shift.Params) {}
```
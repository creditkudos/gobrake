# Airbrake Golang Notifier [![Build Status](https://travis-ci.org/airbrake/gobrake.svg?branch=v2)](https://travis-ci.org/airbrake/gobrake)

<img src="http://f.cl.ly/items/3J3h1L05222X3o1w2l2L/golang.jpg" width=800px>

# Installation
gobrake can be installed like any other go package:

```bash
$ go get github.com/airbrake/gobrake
```

# Example

```go
package main

import (
    "errors"

    "github.com/airbrake/gobrake/v4"
)

var airbrake = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
    ProjectId: 123456,
    ProjectKey: "FIXME",
    Environment: "production",
})

func init() {
    airbrake.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
        notice.Params["user"] = map[string]string{
            "id": "1",
            "username": "johnsmith",
            "name": "John Smith",
        }
        return notice
    })
}

func main() {
    defer airbrake.Close()
    defer airbrake.NotifyOnPanic()

    airbrake.Notify(errors.New("operation failed"), nil)
}
```

## Ignoring notices

```go
airbrake.AddFilter(func(notice *gobrake.Notice) *gobrake.Notice {
    if notice.Context["environment"] == "development" {
        // Ignore notices in development environment.
        return nil
    }
    return notice
})
```

## Setting severity

[Severity](https://airbrake.io/docs/airbrake-faq/what-is-severity/) allows
categorizing how severe an error is. By default, it's set to `error`. To
redefine severity, simply overwrite `context/severity` of a notice object. For
example:

```go
notice := airbrake.NewNotice("operation failed", nil, 0)
notice.Context["severity"] = "critical"
airbrake.Notify(notice, nil)
```

## Logging

You can use [glog fork](https://github.com/airbrake/glog) to send your logs to Airbrake.

## Sending routes stats

In order to collect some basic routes stats you can instrument your application
using `notifier.Routes.Notify` API. We also have prepared HTTP middleware examples for [Gin](examples/gin) and
[Beego](examples/beego).  Here is an example using the net/http middleware.

```go
package main

import (
  "fmt"
  "net/http"

  "github.com/airbrake/gobrake"
)

// Airbrake is used to report errors and track performance
var Airbrake = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
  ProjectId:   123123,              // <-- Fill in this value
  ProjectKey:  "YourProjectAPIKey", // <-- Fill in this value
  Environment: "Production",
})

func indexHandler(w http.ResponseWriter, req *http.Request) {
  fmt.Fprintf(w, "Hello, There!")
}

func main() {
  fmt.Println("Server listening at http://localhost:5555/")
  // Wrap the indexHandler with Airbrake Performance Monitoring middleware:
  http.HandleFunc(airbrakePerformance("/", indexHandler))
  http.ListenAndServe(":5555", nil)
}

func airbrakePerformance(route string, h http.HandlerFunc) (string, http.HandlerFunc) {
  handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    ctx := req.Context()
    ctx, routeMetric := gobrake.NewRouteMetric(ctx, req.Method, route) // Starts the timing
    arw := newAirbrakeResponseWriter(w)

    h.ServeHTTP(arw, req)

    routeMetric.StatusCode = arw.statusCode
    Airbrake.Routes.Notify(ctx, routeMetric) // Stops the timing and reports
    fmt.Printf("code: %v, method: %v, route: %v\n", arw.statusCode, req.Method, route)
  })

  return route, handler
}

type airbrakeResponseWriter struct {
  http.ResponseWriter
  statusCode int
}

func newAirbrakeResponseWriter(w http.ResponseWriter) *airbrakeResponseWriter {
  // Returns 200 OK if WriteHeader isn't called
  return &airbrakeResponseWriter{w, http.StatusOK}
}

func (arw *airbrakeResponseWriter) WriteHeader(code int) {
  arw.statusCode = code
  arw.ResponseWriter.WriteHeader(code)
}
```


To get more detailed timing you can wrap important blocks of code into spans. For example, you can create 2 spans `sql` and `http` to measure timing of specific operations:

``` go
metric := &gobrake.RouteMetric{
    Method: c.Request.Method,
    Route:  routeName,
    StartTime:  time.Now(),
}

metric.StartSpan("sql")
users, err := fetchUser(ctx, userID)
metric.EndSpan("sql")

metric.StartSpan("http")
resp, err := http.Get("http://example.com/")
metric.EndSpan("http")

metric.StatusCode = http.StatusOK
notifier.Routes.Notify(ctx, metric)
```

You can also collect stats about individual SQL queries performance using following API:

```go
notifier.Queries.Notify(&gobrake.QueryInfo{
    Query:     "SELECT * FROM users WHERE id = ?", // query must be normalized
    Func:      "fetchUser", // optional
    File:      "models/user.go", // optional
    Line:      123, // optional
    StartTime: startTime,
    EndTime:   time.Now(),
})
```

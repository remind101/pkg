# Package httpx

Package httpx provides a layer of convenience over "net/http". Specifically:

1. `httpx.Handler`'s return an `error` which makes handler implementations feel
   more idiomatic and reduces the chance of accidentally forgetting to return.

The most important part of package httpx is the `Handler` interface, which is
defined as:

```go
type Handler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request) error
}
```

## Usage

In order to use the `httpx.Handler` interface, you need a compatible router. One is provided within this package that wraps [gorilla mux](https://github.com/gorilla/mux).

```go
r := httpx.NewRouter()
r.HandleFunc("/", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	io.WriteString(w, `ok`)
	return nil
}).Methods("GET")

// Adapt the router to the http.Handler interface and insert a
// context.Background().
s := middleware.Background(r)

http.ListenAndServe(":8080", s)
```

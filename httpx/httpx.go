// package httpx provides an extra layer of convenience over package http.
package httpx

// key used to store context values from within this package.
type key int

const (
	varsKey key = iota
	requestIDKey
	routeKey
)

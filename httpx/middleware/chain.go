package middleware

import (
	"net/http"

	"github.com/remind101/pkg/httpx"
	"golang.org/x/net/context"
)

// Inspiration from https://github.com/justinas/alice

type Link func(httpx.Handler) httpx.Handler

type Chain struct {
	links []Link
}

// NewChain creates a chain of "Link"s.
// Each Link represents a middleware
func NewChain(links ...Link) Chain {
	c := Chain{}
	c.links = append(c.links, links...)

	return c
}

// Prepend a Link to the existing chain
func (c Chain) Prepend(link Link) Chain {
	c.links = append([]Link{link }, c.links...)
	return c
}

// Append a Link to the existing chain
func (c Chain) Append(link Link) Chain {
	c.links = append(c.links, link)
	return c
}

// ThenHandlerFunc is an alias for Then but takes a HandlerFunc instead of an Handler
func (c Chain) ThenHandlerFunc(h func(context.Context, http.ResponseWriter, *http.Request) error) httpx.Handler {
	return c.Then(httpx.HandlerFunc(h))
}

// Then should be called with the final middleware.
func (c Chain) Then(h httpx.Handler) httpx.Handler {
	var final httpx.Handler
	if h != nil {
		final = h
	}

	for i := len(c.links) - 1; i >= 0; i-- {
		final = c.links[i](final)
	}

	return final
}

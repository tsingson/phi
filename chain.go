package phi

import (
	"github.com/valyala/fasthttp"
)

// Chain returns a Middlewares type from a slice of middleware handlers.
func Chain(middlewares ...Middleware) Middlewares {
	return Middlewares(middlewares)
}

// Handler builds and returns a phi.HandlerFunc from the chain of middlewares,
// with `h phi.RequestHandlerFunc` as the final handler.
func (mws Middlewares) Handler(h HandlerFunc) HandlerFunc {
	return &ChainHandler{mws, h, chain(mws, h)}
}

// HandlerFunc builds and returns a phi.HandlerFunc from the chain of middlewares,
// with `h phi.RequestHandlerFunc` as the final handler.
func (mws Middlewares) HandlerFunc(h RequestHandlerFunc) HandlerFunc {
	return &ChainHandler{mws, h, chain(mws, h)}
}

// ChainHandler is a phi.HandlerFunc with support for handler composition and
// execution.
type ChainHandler struct {
	Middlewares Middlewares
	Endpoint    HandlerFunc
	chain       HandlerFunc
}

// Handler return all router as fasthttp.RequestHandler´
func (c *ChainHandler) Handler(ctx *fasthttp.RequestCtx) { // nolint
	c.chain.Handler(ctx)
}

// chain builds a phi.HandlerFunc composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(middlewares Middlewares, endpoint HandlerFunc) HandlerFunc {
	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint.Handler)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

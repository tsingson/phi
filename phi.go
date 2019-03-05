// Package phi is a small, idiomatic and composable router for building HTTP services.
//
// phi requires Go 1.7 or newer.
//
// Example:
//  package main
//
//  import (
//    "log"
//    "time"
//
//    "github.com/fate-lovely/phi"
//    "github.com/valyala/fasthttp"
//  )
//
//  func main() {
//    r := phi.NewRouter()
//
//    reqIDMW := func(next phi.RequestHandlerFunc) phi.RequestHandlerFunc {
//      return func(ctx *fasthttp.RequestCtx) {
//        next(ctx)
//        ctx.WriteString("+reqid=1")
//      }
//    }
//    r.Use(reqIDMW)
//
//    r.Get("/", func(ctx *fasthttp.RequestCtx) {
//      ctx.WriteString("index")
//    })
//    r.NotFound(func(ctx *fasthttp.RequestCtx) {
//      ctx.WriteString("whoops, not found")
//      ctx.SetStatusCode(404)
//    })
//    r.MethodNotAllowed(func(ctx *fasthttp.RequestCtx) {
//      ctx.WriteString("whoops, bad method")
//      ctx.SetStatusCode(405)
//    })
//
//    // tasks
//    r.Group(func(r phi.Router) {
//      mw := func(next phi.RequestHandlerFunc) phi.RequestHandlerFunc {
//        return func(ctx *fasthttp.RequestCtx) {
//          next(ctx)
//          ctx.WriteString("+task")
//        }
//      }
//      r.Use(mw)
//
//      r.Get("/task", func(ctx *fasthttp.RequestCtx) {
//        ctx.WriteString("task")
//      })
//      r.Post("/task", func(ctx *fasthttp.RequestCtx) {
//        ctx.WriteString("new task")
//      })
//
//      caution := func(next phi.RequestHandlerFunc) phi.RequestHandlerFunc {
//        return func(ctx *fasthttp.RequestCtx) {
//          next(ctx)
//          ctx.WriteString("+caution")
//        }
//      }
//      r.With(caution).Delete("/task", func(ctx *fasthttp.RequestCtx) {
//        ctx.WriteString("delete task")
//      })
//    })
//
//    // cat
//    r.Route("/cat", func(r phi.Router) {
//      r.NotFound(func(ctx *fasthttp.RequestCtx) {
//        ctx.WriteString("no such cat")
//        ctx.SetStatusCode(404)
//      })
//      r.Use(func(next phi.RequestHandlerFunc) phi.RequestHandlerFunc {
//        return func(ctx *fasthttp.RequestCtx) {
//          next(ctx)
//          ctx.WriteString("+cat")
//        }
//      })
//      r.Get("/", func(ctx *fasthttp.RequestCtx) {
//        ctx.WriteString("cat")
//      })
//      r.Patch("/", func(ctx *fasthttp.RequestCtx) {
//        ctx.WriteString("patch cat")
//      })
//    })
//
//    // user
//    userRouter := phi.NewRouter()
//    userRouter.NotFound(func(ctx *fasthttp.RequestCtx) {
//      ctx.WriteString("no such user")
//      ctx.SetStatusCode(404)
//    })
//    userRouter.Use(func(next phi.RequestHandlerFunc) phi.RequestHandlerFunc {
//      return func(ctx *fasthttp.RequestCtx) {
//        next(ctx)
//        ctx.WriteString("+user")
//      }
//    })
//    userRouter.Get("/", func(ctx *fasthttp.RequestCtx) {
//      ctx.WriteString("user")
//    })
//    userRouter.Post("/", func(ctx *fasthttp.RequestCtx) {
//      ctx.WriteString("new user")
//    })
//    r.Mount("/user", userRouter)
//
//    server := &fasthttp.Server{
//      HandlerFunc:     r.Handler,
//      ReadTimeout: 10 * time.Second,
//    }
//
//    log.Fatal(server.ListenAndServe(":7789"))
//  }
//
// See github.com/fate-lovely/phi/examples/ for more in-depth examples.
//
package phi

import (
	"github.com/valyala/fasthttp"
)

// HandlerFunc represents a fasthttp request handler,
// it has one method: Handler, which is equal to fasthttp.RequestHandler
type HandlerFunc interface {
	Handler(ctx *fasthttp.RequestCtx)
}

// RequestHandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers.
type RequestHandlerFunc func(ctx *fasthttp.RequestCtx)

// Handler calss fn(ctx)
func (fn RequestHandlerFunc) Handler(ctx *fasthttp.RequestCtx) {
	fn(ctx)
}

// Middleware represents phi middlewares, which accept a RequestHandlerFunc and return a RequestHandlerFunc
type Middleware func(RequestHandlerFunc) RequestHandlerFunc

// Middlewares type is a slice of standard middleware handlers with methods
// to compose middleware chains and phi.HandlerFunc's.
// type Middlewares []func(HandlerFunc) HandlerFunc
type Middlewares []Middleware

// NewRouter returns a new Mux object that implements the Router interface.
func NewRouter() *Mux {
	return NewMux()
}

// Router consisting of the core routing methods used by phi's Mux,
type Router interface {
	HandlerFunc
	Routes

	// Use appends one of more middlewares onto the Router stack.
	Use(middlewares ...Middleware)

	// With adds inline middlewares for an endpoint handler.
	With(middlewares ...Middleware) Router

	// Group adds a new inline-Router along the current routing
	// path, with a fresh middleware stack for the inline-Router.
	Group(fn func(r Router))

	// Route mounts a sub-Router along a `pattern`` string.
	Route(pattern string, fn func(r Router))

	// Mount attaches another phi.HandlerFunc along ./pattern/*
	Mount(pattern string, h HandlerFunc)

	// Handle and HandleFunc adds routes for `pattern` that matches
	// all HTTP methods.
	Handle(pattern string, h RequestHandlerFunc)

	// Method and MethodFunc adds routes for `pattern` that matches
	// the `method` HTTP method.
	Method(method, pattern string, h RequestHandlerFunc)

	// HTTP-method routing along `pattern`
	Connect(pattern string, h RequestHandlerFunc)
	Delete(pattern string, h RequestHandlerFunc)
	Get(pattern string, h RequestHandlerFunc)
	Head(pattern string, h RequestHandlerFunc)
	Options(pattern string, h RequestHandlerFunc)
	Patch(pattern string, h RequestHandlerFunc)
	Post(pattern string, h RequestHandlerFunc)
	Put(pattern string, h RequestHandlerFunc)
	Trace(pattern string, h RequestHandlerFunc)

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	NotFound(h RequestHandlerFunc)

	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
	MethodNotAllowed(h RequestHandlerFunc)
}

// Routes interface adds two methods for router traversal, which is also
// used by the `docgen` subpackage to generation documentation for Routers.
type Routes interface {
	// Routes returns the routing tree in an easily traversable structure.
	Routes() []Route

	// Middlewares returns the list of middlewares in use by the router.
	Middlewares() Middlewares

	// Match searches the routing tree for a handler that matches
	// the method/path - similar to routing a http request, but without
	// executing the handler thereafter.
	Match(rctx *Context, method, path string) bool
}

package phi

import (
	"bytes"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/valyala/fasthttp"
)

func TestMuxBasic(t *testing.T) {
	r := NewRouter()
	h := func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("ok")
	}
	r.Connect("/connect", h)
	r.Delete("/delete", h)
	r.Get("/get", h)
	r.Head("/head", h)
	r.Options("/options", h)
	r.Patch("/patch", h)
	r.Post("/post", h)
	r.Put("/put", h)
	r.Trace("/trace", h)

	r.Method("GET", "/method-get", h)
	r.Handle("/handle", h)

	e := newFastHTTPTester(t, r)
	e.Request("CONNECT", "/connect").Expect().Status(200).Text().Equal("ok")
	e.DELETE("/delete").Expect().Status(200).Text().Equal("ok")
	e.GET("/get").Expect().Status(200).Text().Equal("ok")
	e.HEAD("/head").Expect().Status(200).Text().Equal("ok")
	e.OPTIONS("/options").Expect().Status(200).Text().Equal("ok")
	e.PATCH("/patch").Expect().Status(200).Text().Equal("ok")
	e.POST("/post").Expect().Status(200).Text().Equal("ok")
	e.PUT("/put").Expect().Status(200).Text().Equal("ok")
	e.PUT("/put").Expect().Status(200).Text().Equal("ok")
	e.Request("TRACE", "/trace").Expect().Status(200).Text().Equal("ok")

	e.GET("/method-get").Expect().Status(200).Text().Equal("ok")

	e.Request("CONNECT", "/handle").Expect().Status(200).Text().Equal("ok")
	e.DELETE("/handle").Expect().Status(200).Text().Equal("ok")
	e.GET("/handle").Expect().Status(200).Text().Equal("ok")
	e.HEAD("/handle").Expect().Status(200).Text().Equal("ok")
	e.OPTIONS("/handle").Expect().Status(200).Text().Equal("ok")
	e.PATCH("/handle").Expect().Status(200).Text().Equal("ok")
	e.POST("/handle").Expect().Status(200).Text().Equal("ok")
	e.PUT("/handle").Expect().Status(200).Text().Equal("ok")
	e.PUT("/handle").Expect().Status(200).Text().Equal("ok")
	e.Request("TRACE", "/handle").Expect().Status(200).Text().Equal("ok")
}

func TestMuxURLParams(t *testing.T) {
	r := NewRouter()

	r.Get("/{name}", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString(URLParam(ctx, "name"))
	})
	r.Get("/sub/{name}", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("sub" + URLParam(ctx, "name"))
	})

	e := newFastHTTPTester(t, r)
	e.GET("/hello").Expect().Status(200).Text().Equal("hello")
	e.GET("/hello/all").Expect().Status(404)
	e.GET("/sub/hello").Expect().Status(200).Text().Equal("subhello")
}

func TestMuxUse(t *testing.T) {
	r := NewRouter()
	r.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			ctx.WriteString("+mw1")
		}
	})
	r.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			ctx.WriteString("+mw2")
		}
	})
	r.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("ok")
	})

	e := newFastHTTPTester(t, r)
	e.GET("/").Expect().Status(200).Text().Equal("ok+mw2+mw1")
	e.GET("/nothing").Expect().Status(404).Text().Equal("404 Page not found+mw2+mw1")
}

func TestMuxWith(t *testing.T) {
	r := NewRouter()
	h := func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("ok")
	}
	r.Get("/", h)
	mw := func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			ctx.WriteString("+with")
		}
	}
	r.With(mw).Get("/with", h)

	e := newFastHTTPTester(t, r)
	e.GET("/").Expect().Status(200).Text().Equal("ok")
	e.GET("/with").Expect().Status(200).Text().Equal("ok+with")
}

func TestMuxGroup(t *testing.T) {
	r := NewRouter()
	r.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("index")
	})
	r.Group(func(r Router) {
		r.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx *fasthttp.RequestCtx) {
				next(ctx)
				ctx.WriteString("+group")
			}
		})
		r.Get("/s1", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("s1")
		})
		r.Get("/s2", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("s2")
		})
	})

	e := newFastHTTPTester(t, r)
	e.GET("/").Expect().Status(200).Text().Equal("index")
	e.GET("/s1").Expect().Status(200).Text().Equal("s1+group")
	e.GET("/s2").Expect().Status(200).Text().Equal("s2+group")
}

func TestMuxRoute(t *testing.T) {
	r := NewRouter()
	r.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("index")
	})
	r.Route("/admin", func(r Router) {
		r.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx *fasthttp.RequestCtx) {
				next(ctx)
				ctx.WriteString("+route")
			}
		})
		r.Get("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("admin")
		})
		r.Get("/s1", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("s1")
		})
	})
	e := newFastHTTPTester(t, r)
	e.GET("/").Expect().Status(200).Text().Equal("index")
	e.GET("/admin").Expect().Status(200).Text().Equal("admin+route")
	e.GET("/admin/s1").Expect().Status(200).Text().Equal("s1+route")
}

func TestMuxMount(t *testing.T) {
	r := NewRouter()
	r.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("index")
	})

	sub := NewRouter()
	sub.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			ctx.WriteString("+mount")
		}
	})
	sub.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("admin")
	})
	sub.Get("/s1", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("s1")
	})

	r.Mount("/admin", sub)
	e := newFastHTTPTester(t, r)
	e.GET("/").Expect().Status(200).Text().Equal("index")
	e.GET("/admin").Expect().Status(200).Text().Equal("admin+mount")
	e.GET("/admin/s1").Expect().Status(200).Text().Equal("s1+mount")
}

func TestMuxNotFound(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		r := NewRouter()
		r.NotFound(func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(404)
			ctx.WriteString("not found")
		})
		r.Get("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("ok")
		})
		e := newFastHTTPTester(t, r)
		e.GET("/").Expect().Status(200).Text().Equal("ok")
		e.GET("/no").Expect().Status(404).Text().Equal("not found")
		e.GET("/nono").Expect().Status(404).Text().Equal("not found")
	})

	t.Run("nested", func(t *testing.T) {
		r := NewRouter()
		r.NotFound(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("not found")
			ctx.SetStatusCode(404)
		})

		h := func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("ok")
		}

		// should copy parent NotFound if none
		r.Route("/s1", func(r Router) {
			r.Get("/", h)
		})

		sub := NewRouter()
		sub.NotFound(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("sub not found")
			ctx.SetStatusCode(404)
		})
		sub.Get("/", h)
		r.Mount("/s2", sub)

		e := newFastHTTPTester(t, r)
		e.GET("/no").Expect().Status(404).Text().Equal("not found")
		e.GET("/s1/no").Expect().Status(404).Text().Equal("not found")
		e.GET("/s2/no").Expect().Status(404).Text().Equal("sub not found")
	})
}

func TestMuxMethodNotAllowed(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		r := NewRouter()
		r.MethodNotAllowed(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("bad method")
			ctx.SetStatusCode(405)
		})
		r.Get("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("ok")
		})

		e := newFastHTTPTester(t, r)
		e.GET("/").Expect().Status(200).Text().Equal("ok")
		e.POST("/").Expect().Status(405).Text().Equal("bad method")
	})

	t.Run("nested", func(t *testing.T) {
		r := NewRouter()
		r.Get("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("index")
		})
		r.MethodNotAllowed(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("bad method")
			ctx.SetStatusCode(405)
		})

		// should copy parent MethodNotAllowed if none
		r.Route("/s1", func(r Router) {
			r.Get("/", func(ctx *fasthttp.RequestCtx) {
				ctx.WriteString("s1")
			})
		})

		sub := NewRouter()
		sub.MethodNotAllowed(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("s2 bad method")
			ctx.SetStatusCode(405)
		})
		sub.Get("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("s2")
		})
		r.Mount("/s2", sub)

		e := newFastHTTPTester(t, r)
		e.POST("/").Expect().Status(405).Text().Equal("bad method")
		e.POST("/s1").Expect().Status(405).Text().Equal("bad method")
		e.POST("/s2").Expect().Status(405).Text().Equal("s2 bad method")
	})
}

func TestMuxBigMux(t *testing.T) {
	r := bigMux()
	e := newFastHTTPTester(t, r)

	e.GET("/").Expect().Status(200).Text().Equal("index+reqid=1")
	e.POST("/").Expect().Status(405).Text().Equal("whoops, bad method+reqid=1")
	e.GET("/nothing").Expect().Status(404).Text().Equal("whoops, not found+reqid=1")

	// task
	e.GET("/task").Expect().Status(200).Text().Equal("task+task+reqid=1")
	e.POST("/task").Expect().Status(200).Text().Equal("new task+task+reqid=1")
	e.DELETE("/task").Expect().Status(200).Text().Equal("delete task+caution+task+reqid=1")

	// cat
	e.GET("/cat").Expect().Status(200).Text().Equal("cat+cat+reqid=1")
	e.PATCH("/cat").Expect().Status(200).Text().Equal("patch cat+cat+reqid=1")
	e.GET("/cat/nothing").Expect().Status(404).Text().Equal("no such cat+cat+reqid=1")

	// user
	e.GET("/user").Expect().Status(200).Text().Equal("user+user+reqid=1")
	e.POST("/user").Expect().Status(200).Text().Equal("new user+user+reqid=1")
	e.GET("/user/nothing").Expect().Status(404).Text().Equal("no such user+user+reqid=1")
}

/*----------  Internal  ----------*/

func bigMux() Router {
	r := NewRouter()

	reqIDMW := func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			ctx.WriteString("+reqid=1")
		}
	}
	r.Use(reqIDMW)

	r.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("index")
	})
	r.NotFound(func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("whoops, not found")
		ctx.SetStatusCode(404)
	})
	r.MethodNotAllowed(func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("whoops, bad method")
		ctx.SetStatusCode(405)
	})

	// tasks
	r.Group(func(r Router) {
		mw := func(next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx *fasthttp.RequestCtx) {
				next(ctx)
				ctx.WriteString("+task")
			}
		}
		r.Use(mw)

		r.Get("/task", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("task")
		})
		r.Post("/task", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("new task")
		})

		caution := func(next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx *fasthttp.RequestCtx) {
				next(ctx)
				ctx.WriteString("+caution")
			}
		}
		r.With(caution).Delete("/task", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("delete task")
		})
	})

	// cat
	r.Route("/cat", func(r Router) {
		r.NotFound(func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("no such cat")
			ctx.SetStatusCode(404)
		})
		r.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
			return func(ctx *fasthttp.RequestCtx) {
				next(ctx)
				ctx.WriteString("+cat")
			}
		})
		r.Get("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("cat")
		})
		r.Patch("/", func(ctx *fasthttp.RequestCtx) {
			ctx.WriteString("patch cat")
		})
	})

	// user
	userRouter := NewRouter()
	userRouter.NotFound(func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("no such user")
		ctx.SetStatusCode(404)
	})
	userRouter.Use(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			ctx.WriteString("+user")
		}
	})
	userRouter.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("user")
	})
	userRouter.Post("/", func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("new user")
	})
	r.Mount("/user", userRouter)

	return r
}

func newFastHTTPTester(t *testing.T, h HandlerFunc) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		// Pass requests directly to FastHTTPHandler.
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(fasthttp.RequestHandler(h.Handler)),
			Jar:       httpexpect.NewJar(),
		},
		// Report errors using testify.
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// func TestRouter(t *testing.T) {
// 	r := New()
//
// 	routed := false
// 	r.Method("Get", "/user/:name", func(ctx *fasthttp.RequestCtx) {
// 		routed = true
// 		want := map[string]string{"name": "gopher"}
//
// 		if ctx.UserValue("name") != want["name"] {
// 			t.Fatalf("wrong wildcard values: want %v, got %v", want["name"], ctx.UserValue("name"))
// 		}
// 		ctx.Success("foo/bar", []byte("success"))
// 	})
//
// 	s := &fasthttp.Server{
// 		Handler: r.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	rw.r.WriteString("GET /user/gopher?baz HTTP/1.1\r\n\r\n")
//
// 	ch := make(chan error)
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
//
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
//
// 	if !routed {
// 		t.Fatal("routing failed")
// 	}
// }

type handlerStruct struct {
	handeled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handeled = true
}

func TestRouterAPI(t *testing.T) {
	var get, head, options, post, put, patch, deleted bool

	r := New()
	r.Get("/GET", func(ctx *fasthttp.RequestCtx) {
		get = true
	})
	r.Head("/GET", func(ctx *fasthttp.RequestCtx) {
		head = true
	})
	r.Options("/GET", func(ctx *fasthttp.RequestCtx) {
		options = true
	})
	r.Post("/POST", func(ctx *fasthttp.RequestCtx) {
		post = true
	})
	r.Put("/PUT", func(ctx *fasthttp.RequestCtx) {
		put = true
	})
	r.Patch("/PATCH", func(ctx *fasthttp.RequestCtx) {
		patch = true
	})
	r.Delete("/DELETE", func(ctx *fasthttp.RequestCtx) {
		deleted = true
	})

	s := &fasthttp.Server{
		Handler: r.Handler,
	}

	rw := &readWriter{}
	ch := make(chan error)

	rw.r.WriteString("GET /GET HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !get {
		t.Error("routing GET failed")
	}

	rw.r.WriteString("HEAD /GET HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !head {
		t.Error("routing HEAD failed")
	}

	rw.r.WriteString("OPTIONS /GET HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !options {
		t.Error("routing OPTIONS failed")
	}

	rw.r.WriteString("POST /POST HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !post {
		t.Error("routing POST failed")
	}

	rw.r.WriteString("PUT /PUT HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !put {
		t.Error("routing PUT failed")
	}

	rw.r.WriteString("PATCH /PATCH HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !patch {
		t.Error("routing PATCH failed")
	}

	rw.r.WriteString("DELETE /DELETE HTTP/1.1\r\n\r\n")
	go func() {
		ch <- s.ServeConn(rw)
	}()
	select {
	case err := <-ch:
		if err != nil {
			t.Fatalf("return error %s", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("timeout")
	}
	if !deleted {
		t.Error("routing DELETE failed")
	}
}

func TestRouterRoot(t *testing.T) {
	r := New()
	recv := catchPanic(func() {
		r.Get("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

// func TestRouterChaining(t *testing.T) {
// 	r1 := New()
// 	r2 := New()
// 	//r1.NotFound = r2.Handler
//
// 	fooHit := false
// 	r1.Post("/foo", func(ctx *fasthttp.RequestCtx) {
// 		fooHit = true
// 		ctx.SetStatusCode(fasthttp.StatusOK)
// 	})
//
// 	barHit := false
// 	r2.Post("/bar", func(ctx *fasthttp.RequestCtx) {
// 		barHit = true
// 		ctx.SetStatusCode(fasthttp.StatusOK)
// 	})
//
// 	s := &fasthttp.Server{
// 		Handler: r1.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	ch := make(chan error)
//
// 	rw.r.WriteString("POST /foo HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	br := bufio.NewReader(&rw.w)
// 	var resp fasthttp.Response
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == fasthttp.StatusOK && fooHit) {
// 		t.Errorf("Regular routing failed with router chaining.")
// 		t.FailNow()
// 	}
//
// 	rw.r.WriteString("POST /bar HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == fasthttp.StatusOK && barHit) {
// 		t.Errorf("Chained routing failed with router chaining.")
// 		t.FailNow()
// 	}
//
// 	rw.r.WriteString("POST /qax HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == fasthttp.StatusNotFound) {
// 		t.Errorf("NotFound behavior failed with router chaining.")
// 		t.FailNow()
// 	}
// }

// func TestRouterOPTIONS(t *testing.T) {
// 	// TODO: because fasthttp is not support OPTIONS method now,
// 	// these test cases will be used in the future.
// 	handlerFunc := func(_ *fasthttp.RequestCtx) {}
//
// 	r := New()
// 	r.Post("/path", handlerFunc)
//
// 	// test not allowed
// 	// * (server)
// 	s := &fasthttp.Server{
// 		Handler: r.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	ch := make(chan error)
//
// 	rw.r.WriteString("OPTIONS * HTTP/1.1\r\nHost:\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	br := bufio.NewReader(&rw.w)
// 	var resp fasthttp.Response
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != fasthttp.StatusOK {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
//
// 	// path
// 	rw.r.WriteString("OPTIONS /path HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != fasthttp.StatusOK {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
//
// 	rw.r.WriteString("OPTIONS /doesnotexist HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == fasthttp.StatusNotFound) {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	}
//
// 	// add another method
// 	r.Get("/path", handlerFunc)
//
// 	// test again
// 	// * (server)
// 	rw.r.WriteString("OPTIONS * HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != fasthttp.StatusOK {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
//
// 	// path
// 	rw.r.WriteString("OPTIONS /path HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != fasthttp.StatusOK {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
//
// 	// custom handler
// 	var custom bool
// 	r.Options("/path", func(_ *fasthttp.RequestCtx) {
// 		custom = true
// 	})
//
// 	// test again
// 	// * (server)
// 	rw.r.WriteString("OPTIONS * HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != fasthttp.StatusOK {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, GET, OPTIONS" && allow != "GET, POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
// 	if custom {
// 		t.Error("custom handler called on *")
// 	}
//
// 	// path
// 	rw.r.WriteString("OPTIONS /path HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != fasthttp.StatusOK {
// 		t.Errorf("OPTIONS handling failed: Code=%d, Header=%v",
// 			resp.Header.StatusCode(), resp.Header.String())
// 	}
// 	if !custom {
// 		t.Error("custom handler not called")
// 	}
// }

// func TestRouterNotAllowed(t *testing.T) {
// 	handlerFunc := func(_ *fasthttp.RequestCtx) {}
//
// 	r := New()
// 	r.Post("/path", handlerFunc)
//
// 	// Test not allowed
// 	s := &fasthttp.Server{
// 		Handler: r.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	ch := make(chan error)
//
// 	rw.r.WriteString("GET /path HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	br := bufio.NewReader(&rw.w)
// 	var resp fasthttp.Response
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == fasthttp.StatusMethodNotAllowed) {
// 		t.Errorf("NotAllowed handling failed: Code=%d", resp.Header.StatusCode())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
//
// 	// add another method
// 	r.Delete("/path", handlerFunc)
// 	r.Options("/path", handlerFunc) // must be ignored
//
// 	// test again
// 	rw.r.WriteString("GET /path HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == fasthttp.StatusMethodNotAllowed) {
// 		t.Errorf("NotAllowed handling failed: Code=%d", resp.Header.StatusCode())
// 	} else if allow := string(resp.Header.Peek("Allow")); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
// 		t.Error("unexpected Allow header value: " + allow)
// 	}
//
// 	// responseText := "custom method"
// 	// r.MethodNotAllowed = RequestHandlerFunc(func(ctx *fasthttp.RequestCtx) {
// 	// 	ctx.SetStatusCode(fasthttp.StatusTeapot)
// 	// 	ctx.Write([]byte(responseText))
// 	// })
// 	// rw.r.WriteString("GET /path HTTP/1.1\r\n\r\n")
// 	// go func() {
// 	// 	ch <- s.ServeConn(rw)
// 	// }()
// 	// select {
// 	// case err := <-ch:
// 	// 	if err != nil {
// 	// 		t.Fatalf("return error %s", err)
// 	// 	}
// 	// case <-time.After(100 * time.Millisecond):
// 	// 	t.Fatalf("timeout")
// 	// }
// 	// if err := resp.Read(br); err != nil {
// 	// 	t.Fatalf("Unexpected error when reading response: %s", err)
// 	// }
// 	// if !bytes.Equal(resp.Body(), []byte(responseText)) {
// 	// 	t.Errorf("unexpected response got %q want %q", string(resp.Body()), responseText)
// 	// }
// 	// if resp.Header.StatusCode() != fasthttp.StatusTeapot {
// 	// 	t.Errorf("unexpected response code %d want %d", resp.Header.StatusCode(), fasthttp.StatusTeapot)
// 	// }
// 	// if allow := string(resp.Header.Peek("Allow")); allow != "POST, DELETE, OPTIONS" && allow != "DELETE, POST, OPTIONS" {
// 	// 	t.Error("unexpected Allow header value: " + allow)
// 	// }
// }

// func TestRouterNotFound(t *testing.T) {
// 	handlerFunc := func(_ *fasthttp.RequestCtx) {}
//
// 	r := New()
// 	r.Get("/path", handlerFunc)
// 	r.Get("/dir/", handlerFunc)
// 	r.Get("/", handlerFunc)
//
// 	testRoutes := []struct {
// 		route string
// 		code  int
// 	}{
// 		{"/path/", 301},          // TSR -/
// 		{"/dir", 301},            // TSR +/
// 		{"/", 200},               // TSR +/
// 		{"/PATH", 301},           // Fixed Case
// 		{"/DIR", 301},            // Fixed Case
// 		{"/PATH/", 301},          // Fixed Case -/
// 		{"/DIR/", 301},           // Fixed Case +/
// 		{"/paTh/?name=foo", 301}, // Fixed Case With Params +/
// 		{"/paTh?name=foo", 301},  // Fixed Case With Params +/
// 		{"/../path", 200},        // CleanPath (Not clean by router, this path is cleaned by fasthttp `ctx.Path()`)
// 		{"/nope", 404},           // NotFound
// 	}
//
// 	s := &fasthttp.Server{
// 		Handler: r.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	br := bufio.NewReader(&rw.w)
// 	var resp fasthttp.Response
// 	ch := make(chan error)
// 	for _, tr := range testRoutes {
// 		rw.r.WriteString(fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", tr.route))
// 		go func() {
// 			ch <- s.ServeConn(rw)
// 		}()
// 		select {
// 		case err := <-ch:
// 			if err != nil {
// 				t.Fatalf("return error %s", err)
// 			}
// 		case <-time.After(100 * time.Millisecond):
// 			t.Fatalf("timeout")
// 		}
// 		if err := resp.Read(br); err != nil {
// 			t.Fatalf("Unexpected error when reading response: %s", err)
// 		}
// 		if !(resp.Header.StatusCode() == tr.code) {
// 			t.Errorf("NotFound handling route %s failed: Code=%d want=%d",
// 				tr.route, resp.Header.StatusCode(), tr.code)
// 		}
// 	}
//
// 	// Test custom not found handler
// 	// var notFound bool
// 	// r.NotFound = fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
// 	// 	ctx.SetStatusCode(404)
// 	// 	notFound = true
// 	// })
// 	// rw.r.WriteString("GET /nope HTTP/1.1\r\n\r\n")
// 	// go func() {
// 	// 	ch <- s.ServeConn(rw)
// 	// }()
// 	// select {
// 	// case err := <-ch:
// 	// 	if err != nil {
// 	// 		t.Fatalf("return error %s", err)
// 	// 	}
// 	// case <-time.After(100 * time.Millisecond):
// 	// 	t.Fatalf("timeout")
// 	// }
// 	// if err := resp.Read(br); err != nil {
// 	// 	t.Fatalf("Unexpected error when reading response: %s", err)
// 	// }
// 	// if !(resp.Header.StatusCode() == 404 && notFound == true) {
// 	// 	t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", resp.Header.StatusCode(), string(resp.Header.Peek("Location")))
// 	// }
//
// 	// Test other method than GET (want 307 instead of 301)
// 	r.Patch("/path", handlerFunc)
// 	rw.r.WriteString("PATCH /path/ HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == 307) {
// 		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", resp.Header.StatusCode(), string(resp.Header.Peek("Location")))
// 	}
//
// 	// Test special case where no node for the prefix "/" exists
// 	r = New()
// 	r.Get("/a", handlerFunc)
// 	s.Handler = r.Handler
// 	rw.r.WriteString("GET / HTTP/1.1\r\n\r\n")
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if !(resp.Header.StatusCode() == 404) {
// 		t.Errorf("NotFound handling route / failed: Code=%d", resp.Header.StatusCode())
// 	}
// }

// func TestRouterPanicHandler(t *testing.T) {
// 	r := New()
// 	panicHandled := false
//
// 	r.PanicHandler = func(ctx *fasthttp.RequestCtx, p interface{}) {
// 		panicHandled = true
// 	}
//
// 	r.Handle("PUT", "/user/:name", func(_ *fasthttp.RequestCtx) {
// 		panic("oops!")
// 	})
//
// 	defer func() {
// 		if rcv := recover(); rcv != nil {
// 			t.Fatal("handling panic failed")
// 		}
// 	}()
//
// 	s := &fasthttp.Server{
// 		Handler: r.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	ch := make(chan error)
//
// 	rw.r.WriteString(string("PUT /user/gopher HTTP/1.1\r\n\r\n"))
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(100 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
//
// 	if !panicHandled {
// 		t.Fatal("simulating failed")
// 	}
// }

// func TestRouterLookup(t *testing.T) {
// 	routed := false
// 	wantHandle := func(_ *fasthttp.RequestCtx) {
// 		routed = true
// 	}
//
// 	r := New()
// 	ctx := &fasthttp.RequestCtx{}
//
// 	// try empty router first
// 	handle, tsr := r.Lookup("GET", "/nope", ctx)
// 	if handle != nil {
// 		t.Fatalf("Got handle for unregistered pattern: %v", handle)
// 	}
// 	if tsr {
// 		t.Error("Got wrong TSR recommendation!")
// 	}
//
// 	// insert route and try again
// 	r.GET("/user/:name", wantHandle)
//
// 	handle, _ = r.Lookup("GET", "/user/gopher", ctx)
// 	if handle == nil {
// 		t.Fatal("Got no handle!")
// 	} else {
// 		handle(nil)
// 		if !routed {
// 			t.Fatal("Routing failed!")
// 		}
// 	}
// 	if ctx.UserValue("name") != "gopher" {
// 		t.Error("Param not set!")
// 	}
//
// 	handle, tsr = r.Lookup("GET", "/user/gopher/", ctx)
// 	if handle != nil {
// 		t.Fatalf("Got handle for unregistered pattern: %v", handle)
// 	}
// 	if !tsr {
// 		t.Error("Got no TSR recommendation!")
// 	}
//
// 	handle, tsr = r.Lookup("GET", "/nope", ctx)
// 	if handle != nil {
// 		t.Fatalf("Got handle for unregistered pattern: %v", handle)
// 	}
// 	if tsr {
// 		t.Error("Got wrong TSR recommendation!")
// 	}
// }

// func TestRouterServeFiles(t *testing.T) {
// 	r := New()
//
// 	recv := catchPanic(func() {
// 		r.ServeFiles("/noFilepath", os.TempDir())
// 	})
// 	if recv == nil {
// 		t.Fatal("registering path not ending with '*filepath' did not panic")
// 	}
// 	body := []byte("fake ico")
// 	ioutil.WriteFile(os.TempDir()+"/favicon.ico", body, 0644)
//
// 	r.ServeFiles("/*filepath", os.TempDir())
//
// 	s := &fasthttp.Server{
// 		Handler: r.Handler,
// 	}
//
// 	rw := &readWriter{}
// 	ch := make(chan error)
//
// 	rw.r.WriteString(string("GET /favicon.ico HTTP/1.1\r\n\r\n"))
// 	go func() {
// 		ch <- s.ServeConn(rw)
// 	}()
// 	select {
// 	case err := <-ch:
// 		if err != nil {
// 			t.Fatalf("return error %s", err)
// 		}
// 	case <-time.After(500 * time.Millisecond):
// 		t.Fatalf("timeout")
// 	}
//
// 	br := bufio.NewReader(&rw.w)
// 	var resp fasthttp.Response
// 	if err := resp.Read(br); err != nil {
// 		t.Fatalf("Unexpected error when reading response: %s", err)
// 	}
// 	if resp.Header.StatusCode() != 200 {
// 		t.Fatalf("Unexpected status code %d. Expected %d", resp.Header.StatusCode(), 423)
// 	}
// 	if !bytes.Equal(resp.Body(), body) {
// 		t.Fatalf("Unexpected body %q. Expected %q", resp.Body(), string(body))
// 	}
// }

type readWriter struct {
	net.Conn
	r bytes.Buffer
	w bytes.Buffer
}

var zeroTCPAddr = &net.TCPAddr{
	IP: net.IPv4zero,
}

func (rw *readWriter) Close() error {
	return nil
}

func (rw *readWriter) Read(b []byte) (int, error) {
	return rw.r.Read(b)
}

func (rw *readWriter) Write(b []byte) (int, error) {
	return rw.w.Write(b)
}

func (rw *readWriter) RemoteAddr() net.Addr {
	return zeroTCPAddr
}

func (rw *readWriter) LocalAddr() net.Addr {
	return zeroTCPAddr
}

func (rw *readWriter) SetReadDeadline(t time.Time) error {
	return nil
}

func (rw *readWriter) SetWriteDeadline(t time.Time) error {
	return nil
}

func BenchmarkRouterGet(b *testing.B) {
	resp := []byte("Bench GET")

	r := New()
	r.Get("/bench", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("text/plain", resp)
	})

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/bench")

	for i := 0; i < b.N; i++ {
		r.Handler(ctx)
	}
}

// func BenchmarkRouterNotFound(b *testing.B) {
// 	resp := []byte("Bench Not Found")
//
// 	r := New()
// 	r.Get("/bench", func(ctx *fasthttp.RequestCtx) {
// 		ctx.Success("text/plain", resp)
// 	})
//
// 	ctx := new(fasthttp.RequestCtx)
// 	ctx.Request.Header.SetMethod("GET")
// 	ctx.Request.SetRequestURI("/notfound")
//
// 	for i := 0; i < b.N; i++ {
// 		r.Handler(ctx)
// 	}
// }

func BenchmarkRouterCleanPath(b *testing.B) {
	resp := []byte("Bench GET")

	r := New()
	r.Get("/bench", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("text/plain", resp)
	})

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/../bench/")

	for i := 0; i < b.N; i++ {
		r.Handler(ctx)
	}
}

func BenchmarkRouterRedirectTrailingSlash(b *testing.B) {
	resp := []byte("Bench GET")

	r := New()
	r.Get("/bench/", func(ctx *fasthttp.RequestCtx) {
		ctx.Success("text/plain", resp)
	})

	ctx := new(fasthttp.RequestCtx)
	ctx.Request.Header.SetMethod("GET")
	ctx.Request.SetRequestURI("/bench")

	for i := 0; i < b.N; i++ {
		r.Handler(ctx)
	}
}

func catchPanic(testFunc func()) (recv interface{}) {
	defer func() {
		recv = recover()
	}()

	testFunc()
	return
}

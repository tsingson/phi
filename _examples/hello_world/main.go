package main

import (
	"github.com/oklog/run"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/pprofhandler"
	"github.com/valyala/fasthttp/reuseport"

	"github.com/tsingson/phi"
)

func main() {

	r := phi.New()

	r.Get("/", func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetContentType("text/html; charset=utf-8")
		ctx.SetBodyString("Hello World")
		return
	})

	signal := make(chan struct{})

	//
	var g run.Group

	ln, _ := reuseport.Listen("tcp4", ":8080")
	g.Add(func() error {
		return fasthttp.Serve(ln, r.Handler)
	}, func(error) {
		ln.Close()
	})

	ln1, _ := reuseport.Listen("tcp4", ":8090")
	g.Add(func() error {
		return fasthttp.Serve(ln1, pprofhandler.PprofHandler)
	}, func(error) {
		ln1.Close()
	})

	g.Run()

	<-signal

}

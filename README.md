# <img alt="phi" src="https://cdn.rawgit.com/fate-lovely/phi/master/phi.svg" width="220" />

[![GoDoc Widget]][GoDoc] [![Travis Widget]][Travis] [![License Widget]][License] [![GoReport Widget]][GoReport]

`phi` is a package which ports [chi](https://github.com/go-chi/chi) to fasthttp.

## fork
fork from [fate-lovely/phi](https://github.com/fate-lovely/phi) that [fate-lovely](https://github.com/fate-lovely) port from [chi](https://github.com/go-chi/chi)

my modify:
*  rename func and type name , more clean to read

## Install

`go get -u github.com/fate-lovely/phi`

## Example

```go
r := NewRouter()

reqIDMW := func(next HandlerFunc) HandlerFunc {
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
  mw := func(next HandlerFunc) HandlerFunc {
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

  caution := func(next HandlerFunc) HandlerFunc {
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
  r.Use(func(next HandlerFunc) HandlerFunc {
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
userRouter.Use(func(next HandlerFunc) HandlerFunc {
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
```

output:

|        Path         | Status Code |               Body               |
| :-----------------: | :---------: | :------------------------------: |
|       `GET /`       |     200     |          index+reqid=1           |
|      `POST /`       |     405     |    whoops, bad method+reqid=1    |
|   `GET /nothing`    |     404     |    whoops, not found+reqid=1     |
|     `GET /task`     |     200     |        task+task+reqid=1         |
|    `POST /task`     |     200     |      new task+task+reqid=1       |
|   `DELETE /task`    |     200     | delete task+caution+task+reqid=1 |
|     `GET /cat`      |     200     |         cat+cat+reqid=1          |
|    `PATCH /cat`     |     200     |      patch cat+cat+reqid=1       |
| `GET /cat/nothing`  |     404     |     no such cat+cat+reqid=1      |
|     `GET /user`     |     200     |        user+user+reqid=1         |
|    `POST /user`     |     200     |      new user+user+reqid=1       |
| `GET /user/nothing` |     404     |    no such user+user+reqid=1     |

## License

Licensed under [MIT License](http://mit-license.org/2017)

[License]: http://mit-license.org/2017
[License Widget]: http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square
[GoDoc]: https://godoc.org/github.com/fate-lovely/phi
[GoDoc Widget]: https://godoc.org/github.com/fate-lovely/phi?status.svg
[Travis]: https://travis-ci.org/fate-lovely/phi
[Travis Widget]: https://travis-ci.org/fate-lovely/phi.svg?branch=master
[GoReport Widget]: https://goreportcard.com/badge/github.com/fate-lovely/phi
[GoReport]: https://goreportcard.com/report/github.com/fate-lovely/phi

package pprometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tsingson/phi"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
	"net/url"

	"strconv"
	"sync"
	"time"
)

var (
	defaultMetricPath  = "/metrics"
	requestHandlerPool sync.Pool
)

type FasthttpHandlerFunc func(*fasthttp.RequestCtx)

type Prometheus struct {
	reqCnt            *prometheus.CounterVec
	reqDur            *prometheus.HistogramVec
	reqSize, respSize prometheus.Summary
	router            *phi.Mux

	MetricsPath string
}

func NewPrometheus(subsystem string) *Prometheus {

	p := &Prometheus{
		MetricsPath: defaultMetricPath,
	}
	p.registerMetrics(subsystem)

	return p
}

func prometheusHandler() phi.HandlerFunc {
	return promeHndler(promhttp.Handler())
}



func (p *Prometheus) WrapHandler(r *phi.Mux) fasthttp.RequestHandler {

	// Setting prometheus metrics handler
	r.Get(p.MetricsPath, prometheusHandler())

	return func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Request.URI().Path()) == defaultMetricPath {
			r.ServeFastHTTP(ctx)
			return
		}

		reqSize := make(chan int)
		frc := acquireRequestFromPool()
		ctx.Request.CopyTo(frc)
		go computeApproximateRequestSize(frc, reqSize)

		start := time.Now()
		r.ServeFastHTTP(ctx)

		status := strconv.Itoa(ctx.Response.StatusCode())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		respSize := float64(len(ctx.Response.Body()))

		p.reqDur.WithLabelValues(status).Observe(elapsed)
		p.reqCnt.WithLabelValues(status, string(ctx.Method())).Inc()
		p.reqSize.Observe(float64(<-reqSize))
		p.respSize.Observe(respSize)
	}
}

// Idea is from https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go and https://github.com/zsais/go-gin-prometheus/blob/master/middleware.go
func computeApproximateRequestSize(ctx *fasthttp.Request, out chan int) {
	s := 0
	if ctx.URI() != nil {
		s += len(ctx.URI().Path())
		s += len(ctx.URI().Host())
	}

	s += len(ctx.Header.Method())
	s += len("HTTP/1.1")

	ctx.Header.VisitAll(func(key, value []byte) {
		if string(key) != "Host" {
			s += len(key) + len(value)
		}
	})

	if ctx.Header.ContentLength() != -1 {
		s += ctx.Header.ContentLength()
	}

	out <- s
}

func (p *Prometheus) registerMetrics(subsystem string) {

	RequestDurationBucket := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60}

	p.reqCnt = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "The HTTP request counts processed.",
		},
		[]string{"code", "method"},
	)

	p.reqDur = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request duration in seconds.",
			Buckets:   RequestDurationBucket,
		},
		[]string{"code"},
	)

	p.reqSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
	)

	p.respSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
	)

	prometheus.MustRegister(p.reqCnt, p.reqDur, p.reqSize, p.respSize)
}

func acquireRequestFromPool() *fasthttp.Request {
	rp := requestHandlerPool.Get()

	if rp == nil {
		return new(fasthttp.Request)
	}

	frc := rp.(*fasthttp.Request)
	return frc
}
//

func promeHndler(h http.Handler) phi.HandlerFunc {

	return func(ctx *fasthttp.RequestCtx) {
		var r http.Request

		body := ctx.PostBody()
		r.Method = string(ctx.Method())
		r.Proto = "HTTP/1.1"
		r.ProtoMajor = 1
		r.ProtoMinor = 1
		r.RequestURI = string(ctx.RequestURI())
		r.ContentLength = int64(len(body))
		r.Host = string(ctx.Host())
		r.RemoteAddr = ctx.RemoteAddr().String()

		hdr := make(http.Header)
		ctx.Request.Header.VisitAll(func(k, v []byte) {
			sk := string(k)
			sv := string(v)
			switch sk {
			case "Transfer-Encoding":
				r.TransferEncoding = append(r.TransferEncoding, sv)
			default:
				hdr.Set(sk, sv)
			}
		})
		r.Header = hdr
		r.Body = &netHTTPBody{body}
		rURL, err := url.ParseRequestURI(r.RequestURI)
		if err != nil {
			ctx.Logger().Printf("cannot parse requestURI %q: %s", r.RequestURI, err)
			ctx.Error("Internal Server Error", fasthttp.StatusInternalServerError)
			return
		}
		r.URL = rURL

		var w netHTTPResponseWriter
		h.ServeHTTP(&w, &r)

		ctx.SetStatusCode(w.StatusCode())
		for k, vv := range w.Header() {
			for _, v := range vv {
				ctx.Response.Header.Set(k, v)
			}
		}
		ctx.Write(w.body)
	}
}


type netHTTPBody struct {
	b []byte
}

func (r *netHTTPBody) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func (r *netHTTPBody) Close() error {
	r.b = r.b[:0]
	return nil
}

type netHTTPResponseWriter struct {
	statusCode int
	h          http.Header
	body       []byte
}

func (w *netHTTPResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *netHTTPResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}

func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}

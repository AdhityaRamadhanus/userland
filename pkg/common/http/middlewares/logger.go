package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
	log "github.com/sirupsen/logrus"
)

type logMetricResponseWriter struct {
	http.ResponseWriter
	RequestCount   metrics.Counter
	RequestLatency metrics.Histogram
	StatusCode     int
}

func (lrw *logMetricResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

var (
	LogMetricKeys = []string{"endpoint", "status"}
)

//LogRequest with info level every http request, unless production
func LogMetricRequest(counter metrics.Counter, latencyObserver metrics.Histogram) Middleware {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			lrw := &logMetricResponseWriter{
				ResponseWriter: res,
				StatusCode:     http.StatusOK,
				RequestCount:   counter,
				RequestLatency: latencyObserver,
			}
			nextHandler.ServeHTTP(lrw, req)

			log.WithFields(log.Fields{
				"request_id": req.Header.Get("X-Request-ID"),
				"status":     lrw.StatusCode,
				"method":     req.Method,
				"resp time":  time.Since(start),
			}).Info("PATH " + req.URL.Path)

			statusStr := strconv.Itoa(lrw.StatusCode)
			lrw.RequestCount.With("endpoint", req.URL.Path, "status", statusStr).Add(1)
			lrw.RequestLatency.With("endpoint", req.URL.Path, "status", statusStr).Observe(time.Since(start).Seconds())
		})
	}
}

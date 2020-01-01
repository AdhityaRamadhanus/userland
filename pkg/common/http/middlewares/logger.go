package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/sirupsen/logrus"
)

type logMetricResponseWriter struct {
	http.ResponseWriter
	RequestLatency metrics.Histogram
	StatusCode     int
}

func (lrw *logMetricResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

var (
	LogMetricKeys = []string{"method", "route", "status_code"}
)

//LogRequest with info level every http request, unless production
func LogMetricRequest(latencyObserver metrics.Histogram) Middleware {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			lrw := &logMetricResponseWriter{
				ResponseWriter: res,
				StatusCode:     http.StatusOK,
				RequestLatency: latencyObserver,
			}
			nextHandler.ServeHTTP(lrw, req)

			duration := time.Since(start)
			logrus.WithFields(logrus.Fields{
				"request_id": req.Header.Get("X-Request-ID"),
				"status":     lrw.StatusCode,
				"method":     req.Method,
				"resp time":  duration,
			}).Info("PATH " + req.URL.Path)

			statusStr := strconv.Itoa(lrw.StatusCode)
			lrw.RequestLatency.With("method", req.Method, "route", req.URL.Path, "status_code", statusStr).Observe(duration.Seconds())
		})
	}
}

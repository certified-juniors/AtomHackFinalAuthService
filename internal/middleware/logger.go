package middleware

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type AccessLogger struct {
	LogrusLogger *logrus.Logger
}

func NewLogger(lg *logrus.Logger) *AccessLogger {
	return &AccessLogger{LogrusLogger: lg}
}

func (ac *AccessLogger) AccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)

		ac.LogrusLogger.WithFields(logrus.Fields{
			"method":      r.Method,
			"remote_addr": r.RemoteAddr,
			"request_id":  r.Header.Get("Request-ID"),
			"work_time":   time.Since(start),
			"status":      lrw.statusCode,
		}).Info(r.URL.Path)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

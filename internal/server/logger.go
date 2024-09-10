package server

import (
	"net/http"
	"time"

	"prayago-metricsalert/internal/logger"
)

type (
	responseStats struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		stats *responseStats
	}
)

func (resp *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := resp.ResponseWriter.Write(b)
	if err == nil {
		resp.stats.size += size
	}

	return size, err
}

func (resp *loggingResponseWriter) WriteHeader(statusCode int) {
	resp.ResponseWriter.WriteHeader(statusCode)
	resp.stats.status = statusCode
}

func HTTPHandlerWithLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(respW http.ResponseWriter, req *http.Request) {
		start := time.Now()

		respStats := &responseStats{
			status: 0,
			size:   0,
		}
		respWriter := loggingResponseWriter{
			ResponseWriter: respW,
			stats:          respStats,
		}
		next.ServeHTTP(&respWriter, req)
		duration := time.Since(start)

		logger.LogSugar.Infoln(
			"uri", req.RequestURI,
			"method", req.Method,
			"duration", duration,
			"resp status", respStats.status,
			"resp size", respStats.size,
		)
	})
}

package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()
var LogSugar *zap.SugaredLogger

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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err == nil {
		r.stats.size += size
	}

	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.stats.status = statusCode
}

func init() {
	logger, _ := zap.NewProduction()
	Log = logger
	defer logger.Sync()

	LogSugar = logger.Sugar()
}

func HTTPHandlerWithLogger(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		respStats := &responseStats{
			status: 0,
			size:   0,
		}
		respWriter := loggingResponseWriter{
			ResponseWriter: w,
			stats:          respStats,
		}
		h.ServeHTTP(&respWriter, r)
		duration := time.Since(start)

		LogSugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
			"resp status", respStats.status,
			"resp size", respStats.size,
		)
	}
}

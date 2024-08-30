package logger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	fmt.Println("status", statusCode)
	r.responseData.status = statusCode
}

func LoggingWrapper(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		logger, err := zap.NewDevelopment()
		if err != nil {
			log.Printf("error ocured while creating zap logger: %v", err)
		}
		defer logger.Sync()

		sugar := *logger.Sugar()

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", time.Since(start),
		)

		h.ServeHTTP(&lw, r)

		sugar.Infoln(
			"status", responseData.status,
			"size", responseData.size,
			"duration", time.Since(start),
		)
	}

	return http.HandlerFunc(logFn)
}

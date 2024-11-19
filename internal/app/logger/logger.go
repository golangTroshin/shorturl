// Package logger provides middleware for HTTP logging.
//
// It includes functionality to log HTTP request and response data such as status codes,
// sizes, methods, and durations using the zap structured logging library.
package logger

import (
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// responseData is a structure that stores metadata about an HTTP response.
	//
	// Fields:
	//   - status: The HTTP status code returned in the response.
	//   - size  : The size of the response body in bytes.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter is a custom implementation of `http.ResponseWriter`.
	// It wraps the standard `http.ResponseWriter` to capture additional response data for logging.
	//
	// Fields:
	//   - ResponseWriter: The original `http.ResponseWriter` to delegate writes to.
	//   - responseData  : A pointer to `responseData` that stores response metadata.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write captures the size of the response body written and delegates to the original `ResponseWriter`.
//
// Parameters:
//   - b: The byte slice to write.
//
// Returns:
//   - The number of bytes written and any error encountered.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the status code and delegates to the original `ResponseWriter`.
//
// Parameters:
//   - statusCode: The HTTP status code to set in the response.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	log.Println("status", statusCode)
	r.responseData.status = statusCode
}

// LoggingWrapper is middleware for logging HTTP requests and responses.
//
// It logs the following data for each request:
//   - The URI of the request.
//   - The HTTP method used.
//   - The duration of the request processing.
//   - The status code returned.
//   - The size of the response body.
//
// It uses the zap structured logging library to log this data.
//
// Parameters:
//   - h: The `http.Handler` to wrap.
//
// Returns:
//   - An `http.Handler` that logs HTTP request and response metadata.
func LoggingWrapper(h http.Handler) http.Handler {
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

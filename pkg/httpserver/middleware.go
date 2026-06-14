package httpserver

import (
	"fmt"
	"net/http"
	"time"

	"go_server/pkg/logging"
)

// responseRecorder wraps http.ResponseWriter so GoServer can track
// the final status code and the number of bytes written to the response.
type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

// newResponseRecorder creates a recorder with a safe default status.
// If WriteHeader is never called explicitly, HTTP assumes 200 OK.
func newResponseRecorder(ResponseWriter http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: ResponseWriter,
		statusCode:     http.StatusOK,
		bytesWritten:   0,
	}
}

// WriteHeader captures the outgoing status code before passing it through.
func (ResponseRecorder *responseRecorder) WriteHeader(StatusCode int) {
	ResponseRecorder.statusCode = StatusCode
	ResponseRecorder.ResponseWriter.WriteHeader(StatusCode)
}

// Write captures the response size.
// It also preserves the default 200 status behavior if WriteHeader
// was not explicitly called earlier.
func (ResponseRecorder *responseRecorder) Write(Data []byte) (int, error) {
	BytesCount, err := ResponseRecorder.ResponseWriter.Write(Data)
	ResponseRecorder.bytesWritten += BytesCount
	return BytesCount, err
}

// MethodMiddleware ensures the incoming HTTP request uses the correct method (GET, POST, etc.).
// This adds a layer of security by rejecting unexpected request types immediately.
func MethodMiddleware(AllowedMethod string, NextHandler http.Handler) http.Handler {
	// middleware logger instance
	MiddlewareLogger := logging.Get("GoServer-Middleware")

	return http.HandlerFunc(func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		if Request.Method != AllowedMethod {
			MiddlewareLogger.Warn("Method Not Allowed",
				"path", Request.URL.Path,
				"received", Request.Method,
				"expected", AllowedMethod,
			)

			// RFC 7231 requires sending the 'Allow' header with a 405 response
			ResponseWriter.Header().Set("Allow", AllowedMethod)
			http.Error(ResponseWriter, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Method is correct, proceed to the actual logic
		NextHandler.ServeHTTP(ResponseWriter, Request)
	})
}

// LoggingMiddleware records the details of every request, including how long it took to process.
// take in the http.Handler so it can wrap both HandlerFuncs and FileServers.
func LoggingMiddleware(NextHandler http.Handler) http.Handler {
	MiddlewareLogger := logging.Get("GoServer-Traffic")

	return http.HandlerFunc(func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		StartTime := time.Now()
		ResponseRecorder := newResponseRecorder(ResponseWriter)

		// Log the incoming request
		MiddlewareLogger.Info("Request Received",
			"method", Request.Method,
			"path", Request.URL.Path,
			"remote", Request.RemoteAddr,
		)

		// Process the request
		NextHandler.ServeHTTP(ResponseRecorder, Request)

		// Log the completion and duration
		MiddlewareLogger.Info("Request Processed",
			"method", Request.Method,
			"path", Request.URL.Path,
			"status", ResponseRecorder.statusCode,
			"bytes", ResponseRecorder.bytesWritten,
			"duration", time.Since(StartTime),
		)
	})
}

// RecoveryMiddleware catches panics within handlers to prevent the entire
// GoServer from crashing. It logs the error and returns a 500 status code.
func RecoveryMiddleware(GoServer *GoServer, NextHandler http.Handler) http.Handler {
	// We also get a specific logger for the recovery process itself.
	RecoveryLogger := logging.Get("GoServer-Recovery")

	return http.HandlerFunc(func(ResponseWriter http.ResponseWriter, Request *http.Request) {
		defer func() {
			if CapturedPanic := recover(); CapturedPanic != nil {
				// 1. LOG THE PANIC IMMEDIATELY
				// This captures the raw panic value before we process the error page.
				RecoveryLogger.Error("CRITICAL PANIC DETECTED",
					"panic_value", CapturedPanic,
					"request_path", Request.URL.Path,
				)

				// 2. RENDER THE INTERNAL ERROR PAGE
				// This call will ALSO log the error via RenderGoServerError.
				GoServer.RenderGoServerError(ResponseWriter, GoServerError{
					StatusCode:   http.StatusInternalServerError,
					Title:        "System Interruption",
					Message:      "We've encountered a critical error. The system recovered, but the request could not be completed.",
					TechnicalErr: fmt.Sprintf("Panic: %v", CapturedPanic),
				})
			}
		}()

		NextHandler.ServeHTTP(ResponseWriter, Request)
	})
}

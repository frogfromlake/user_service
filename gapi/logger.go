package gapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GrpcLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	startTime := time.Now()
	result, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.Unknown
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	logger := log.Info()
	if err != nil {
		logger = log.Error().Err(err)
	}

	// Check if the request is a gRPC health check
	if strings.Contains(info.FullMethod, "Health/Check") {
		// Log only if the status code is not OK
		if err != nil || statusCode != codes.OK {
			logger.Str("protocol", "grpc").
				Str("method", info.FullMethod).
				Int("status_code", int(statusCode)).
				Dur("duration", duration).
				Msg("Received grpc request")
		}
	} else {
		// Log all other requests
		logger.Str("protocol", "grpc").
			Str("method", info.FullMethod).
			Int("status_code", int(statusCode)).
			Str("status_text", statusCode.String()).
			Dur("duration", duration).
			Msg("Received grpc request")
	}

	return result, err
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       []byte
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(body []byte) (int, error) {
	r.Body = body
	return r.ResponseWriter.Write(body)
}

func HttpLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		startTime := time.Now()
		rec := &ResponseRecorder{
			ResponseWriter: res,
			StatusCode:     http.StatusOK,
		}
		handler.ServeHTTP(rec, req)
		duration := time.Since(startTime)

		logger := log.Info()
		if rec.StatusCode != http.StatusOK {
			logger = log.Error().Bytes("body", rec.Body)
		}

		// Check if the request is a health check
		if req.URL.Path == "/streamfair/v1/healthz" {
			// Log only if the status code is not OK
			if rec.StatusCode != http.StatusOK {
				logger.Str("protocol", "http").
					Str("method", req.Method).
					Str("path", req.RequestURI).
					Int("status_code", rec.StatusCode).
					Str("status_text", http.StatusText(rec.StatusCode)).
					Dur("duration", duration).
					Msg("Received HTTP request")
			}
		} else {
			// Log all other requests
			logger.Str("protocol", "http").
				Str("method", req.Method).
				Str("path", req.RequestURI).
				Int("status_code", rec.StatusCode).
				Str("status_text", http.StatusText(rec.StatusCode)).
				Dur("duration", duration).
				Msg("Received HTTP request")
		}
	})
}

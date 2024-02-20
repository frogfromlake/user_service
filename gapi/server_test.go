package gapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestGRPCServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_db.NewMockStore(ctrl)

	// Start the gRPC server
	server := newTestServer(t, mockStore)
	defer server.Shutdown()

	// Start the gRPC server in a goroutine
	go func() {
		server.RunGrpcServer()
	}()

	// Use a context with a deadline to wait for the server to start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for the server to become ready
	if err := waitForServer(ctx, server.config.GrpcServerAddress); err != nil {
		t.Fatalf("failed to wait for server: %v", err)
	}

	// Perform the gRPC health check
	conn, err := grpc.DialContext(ctx, server.config.GrpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial server: %v", err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatalf("failed to perform health check: %v", err)
	}

	// Assert the health check response
	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected status %s, got %s", grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
	}
}

func TestGRPCGatewayServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	server := newTestServer(t, mock_db.NewMockStore(ctrl))

	// Start the gRPC gateway server in a goroutine
	go func() {
		server.RunGrpcGatewayServer()
	}()

	// Use a context with a deadline to wait for the server to start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Wait for the server to become ready
	if err := waitForHTTPServer(ctx, server.config.HttpServerAddress); err != nil {
		t.Fatalf("failed to wait for server: %v", err)
	}

	// Perform the HTTP health check
	url := fmt.Sprintf("http://%s/v1/healthz", server.config.HttpServerAddress)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request to server: %v", err)
	}
	defer resp.Body.Close()

	// Read the body of the response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Assert the health check response
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "OK", string(bodyBytes))
}

// Helper function to wait for the server to become ready
func waitForServer(ctx context.Context, address string) error {
	maxAttempts := 10
	attemptInterval := time.Second

	for i := 0; i < maxAttempts; i++ {
		conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(attemptInterval)
	}
	return fmt.Errorf("server not ready at address %s", address)
}

func waitForHTTPServer(ctx context.Context, address string) error {
	maxAttempts :=  10
	attemptInterval := time.Second

	for i :=  0; i < maxAttempts; i++ {
		resp, err := http.Get(fmt.Sprintf("http://%s/v1/healthz", address))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		time.Sleep(attemptInterval)
	}
	return fmt.Errorf("HTTP server not ready at address %s", address)
}

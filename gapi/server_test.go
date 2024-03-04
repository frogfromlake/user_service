package gapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestGRPCServer(t *testing.T) {
	// Use a sub-test to isolate the environment and ensure cleanup is executed
	t.Run("gRPC Server Test", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mock_db.NewMockStore(ctrl)

		// Start the gRPC server
		server := newTestServer(t, mockStore)
		defer server.Shutdown()

		// Start the gRPC server in a goroutine
		go server.RunGrpcServer()

		// Use a context with a deadline to wait for the server to start
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		tlsConfig, err := LoadTLSConfigWithTrustedCerts(server.config.CertPem, server.config.KeyPem, server.config.CaCertPem)
		require.NoError(t, err)

		creds := credentials.NewTLS(tlsConfig)

		// Wait for the server to become ready
		if err := waitForServer(ctx, server.config.GrpcServerAddress, creds); err != nil {
			t.Fatalf("failed to wait for server: %v", err)
		}

		// Perform the gRPC health check
		conn, err := grpc.DialContext(ctx, server.config.GrpcServerAddress, grpc.WithTransportCredentials(creds))
		require.NoError(t, err)

		defer conn.Close()

		client := grpc_health_v1.NewHealthClient(conn)
		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
		require.NoError(t, err)

		// Assert the health check response
		if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			t.Fatalf("expected status %s, got %s", grpc_health_v1.HealthCheckResponse_SERVING, resp.GetStatus())
		}
	})
}

// Helper function to wait for the server to become ready
func waitForServer(ctx context.Context, address string, creds credentials.TransportCredentials) error {
	maxAttempts := 10
	attemptInterval := time.Second

	for i := 0; i < maxAttempts; i++ {
		conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(creds))
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(attemptInterval)
	}
	return fmt.Errorf("server not ready at address %s", address)
}

func TestGRPCGatewayServer(t *testing.T) {
	// Use a sub-test to isolate the environment and ensure cleanup is executed
	t.Run("gRPC Gateway Server Test", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockStore := mock_db.NewMockStore(ctrl)

		server := newTestServer(t, mockStore)
		defer server.Shutdown()

		config, err := util.LoadConfig()
		require.NoError(t, err)

		// Start the gRPC server and the grpc gateway server in goroutines
		go server.RunGrpcServer()
		go server.RunGrpcGatewayServer()

		// Use a context with a deadline to wait for the server to start
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Wait for the server to become ready
		if err := waitForHTTPServer(server.config); err != nil {
			t.Fatalf("failed to wait for server: %v", err)
		}

		tlsConfig, err := LoadTLSConfigWithTrustedCerts(config.CertPem, config.KeyPem, config.CaCertPem)
		require.NoError(t, err)
		// Create a custom HTTP client that trusts the server's certificate
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}

		// Perform the HTTP health check using the custom client
		url := fmt.Sprintf("https://%s/streamfair/v1/healthz", config.HttpServerAddress)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		resp, err := client.Do(req)
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
		require.Equal(t, "{\"status\":\"SERVING\"}\n", string(bodyBytes))
	})
}

func waitForHTTPServer(config util.Config) error {
	var caCert []byte
	var err error

	maxAttempts := 10
	attemptInterval := time.Second

	if viper.GetString("CI") != "true" {
		// Load  CA certificate
		caCert, err = os.ReadFile(config.CaCertPem)
		if err != nil {
			return fmt.Errorf("failed to load CA certificate: %w", err)
		}
	} else {
		caCert = []byte(config.CaCertPem)
	}

	// Create a new certificate pool and add the CA certificate
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to append CA certificate to pool")
	}

	// Create a custom HTTP client that trusts the CA certificate
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(fmt.Sprintf("https://%s/streamfair/v1/healthz", config.HttpServerAddress))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		time.Sleep(attemptInterval)
	}
	return fmt.Errorf("HTTP server not ready at address %s", config.HttpServerAddress)
}

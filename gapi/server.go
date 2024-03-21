package gapi

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"os"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	_ "github.com/Streamfair/streamfair_user_svc/doc/statik"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

// Server serves gRPC requests for the streamfair user management service.
type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	pb.UnimplementedUserServiceServer
	config          util.Config
	store           db.Store
	healthSrv       *health.Server
	localTokenMaker token.Maker
}

// NewServer creates a new gRPC server.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	localTokenMaker, err := token.NewLocalPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create local token maker: %w", err)
	}

	tlsConfig, err := LoadTLSConfigWithTrustedCerts(config.CertPem, config.KeyPem, config.CaCertPem)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config for 'NewServer': %w", err)
	}

	creds := credentials.NewTLS(tlsConfig)
	grpcLogger := grpc.UnaryInterceptor(GrpcLogger)

	server := &Server{
		config:          config,
		store:           store,
		grpcServer:      grpc.NewServer(grpc.Creds(creds), grpcLogger),
		httpServer:      &http.Server{},
		healthSrv:       health.NewServer(),
		localTokenMaker: localTokenMaker,
	}

	grpc_health_v1.RegisterHealthServer(server.grpcServer, server.healthSrv)

	return server, nil
}

// RunGrpcServer: runs a gRPC server on the given address.
func (server *Server) RunGrpcServer() {
	pb.RegisterUserServiceServer(server.grpcServer, server)
	reflection.Register(server.grpcServer)

	server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", server.config.GrpcServerAddress)
	if err != nil {
		server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Fatal().Err(err).Msg("server: error while creating gRPC listener:")
	}

	log.Info().Msgf("start gRPC server on %s", listener.Addr().String())
	if err := server.grpcServer.Serve(listener); err != nil {
		server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Fatal().Err(err).Msg("server: error while serving gRPC:")
	}
}

// RunGrpcGatewayServer: runs a gRPC gateway server that translates HTTP requests into gRPC calls.
func (server *Server) RunGrpcGatewayServer() {
	tlsConfig, err := LoadTLSConfigWithTrustedCerts(server.config.CertPem, server.config.KeyPem, server.config.CaCertPem)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load TLS config:")
	}

	healthClient, err := CreateHealthClient(context.Background(), server.config.GrpcServerAddress, tlsConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create health client:")
	}

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithHealthEndpointAt(healthClient, "/streamfair/v1/healthz"),
	)

	if err := pb.RegisterUserServiceHandlerServer(context.Background(), grpcMux, server); err != nil {
		log.Fatal().Err(err).Msg("server: error while registering gRPC server:")
	}

	// Add the HTTP logger middleware
	httpLogger := HttpLogger(grpcMux)

	mux := http.NewServeMux()
	mux.Handle("/", httpLogger)

	if err := ServeSwaggerUI(mux); err != nil {
		log.Fatal().Err(err).Msg("Failed to serve Swagger UI:")
	}

	handler := h2c.NewHandler(mux, &http2.Server{})
	server.httpServer.Handler = handler

	if err := StartHTTPServer(server.httpServer, server.config, server.config.CertPem, server.config.KeyPem); err != nil {
		log.Fatal().Err(err).Msg("Failed to start HTTP server:")
	}
}

// LoadTLSConfigWithTrustedCerts loads the TLS configuration either from the specified paths
// or using raw PEM data, depending on the CI environment variable.
func LoadTLSConfigWithTrustedCerts(certPath, keyPath, caCertPath string) (*tls.Config, error) {
	var cert tls.Certificate
	var err error

	// Check if CI is set to true
	if viper.GetString("CI") == "true" {
		// Load the server's certificate and private key from raw PEM data
		cert, err = tls.X509KeyPair([]byte(certPath), []byte(keyPath))
	} else {
		// Load the server's certificate and private key from paths
		cert, err = tls.LoadX509KeyPair(certPath, keyPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificates: %w", err)
	}

	var caCertPEM []byte
	if viper.GetString("CI") == "true" {
		// Use raw PEM data for the CA's certificate
		caCertPEM = []byte(caCertPath)
	} else {
		// Load the CA's certificate from a path
		caCertPEM, err = os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
		}
	}

	// Create a new certificate pool and add the CA's certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPEM) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	return &tls.Config{
		RootCAs:      certPool,                // Use the CA's certificate pool
		Certificates: []tls.Certificate{cert}, // Use the server's certificate and key
	}, nil
}

// CreateHealthClient creates a gRPC health client to be used for health checks.
func CreateHealthClient(ctx context.Context, grpcServerAddress string, tlsConfig *tls.Config) (grpc_health_v1.HealthClient, error) {
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.DialContext(ctx, grpcServerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to dial health server: %w", err)
	}

	return grpc_health_v1.NewHealthClient(conn), nil
}

// ServeSwaggerUI configures the HTTP server to serve the Swagger UI files.
func ServeSwaggerUI(mux *http.ServeMux) error {
	statikFS, err := fs.New()
	if err != nil {
		return fmt.Errorf("error while creating statik file system: %w", err)
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	return nil
}

// StartHTTPServer starts the HTTP server with TLS enabled.
func StartHTTPServer(server *http.Server, config util.Config, certPath, keyPath string) error {
	if viper.GetString("CI") == "true" {
		tlsConfig, err := LoadTLSConfigWithTrustedCerts(config.CertPem, config.KeyPem, config.CaCertPem)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load TLS config:")
		}

		// Set the TLSConfig on the http.Server
		server.TLSConfig = tlsConfig
		certPath = ""
		keyPath = ""
	}

	listener, err := net.Listen("tcp", config.HttpServerAddress)
	if err != nil {
		return fmt.Errorf("error while creating HTTP listener: %w", err)
	}

	log.Info().Msgf("start HTTP Gateway server on %s", listener.Addr().String())
	if err := server.ServeTLS(listener, certPath, keyPath); err != nil {
		return fmt.Errorf("error while starting HTTP Gateway server: %w", err)
	}

	return nil
}

func (server *Server) Shutdown() {
	server.grpcServer.GracefulStop()
}

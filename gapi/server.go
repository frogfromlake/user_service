package gapi

import (
	"context"
	"log"
	"net"
	"net/http"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	_ "github.com/Streamfair/streamfair_user_svc/doc/statik"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
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
	server := &Server{
		config:     config,
		store:      store,
		grpcServer: grpc.NewServer(),
		httpServer: &http.Server{Handler: nil},
		healthSrv:  health.NewServer(),
	}

	grpc_health_v1.RegisterHealthServer(server.grpcServer, server.healthSrv)

	return server, nil
}

// RunGrpcServer runs a gRPC server on the given address.
func (server *Server) RunGrpcServer() {
	pb.RegisterUserServiceServer(server.grpcServer, server)
	reflection.Register(server.grpcServer)

	// Set the initial health status to SERVING when the server starts.
	server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", server.config.GrpcServerAddress)
	if err != nil {
		// Update the health status to NOT_SERVING if there's an error.
		server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Fatalf("server: error while creating gRPC listener: %v", err)
	}

	log.Printf("start gRPC server on %s", listener.Addr().String())
	err = server.grpcServer.Serve(listener)
	if err != nil {
		// Update the health status to NOT_SERVING if there's an error.
		server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		log.Fatalf("server: error while serving gRPC: %v", err)
	}
}

// RunGrpcGatewayServer runs a gRPC gateway server that translates HTTP requests into gRPC calls.
func (server *Server) RunGrpcGatewayServer() {
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})

	grpcMux := runtime.NewServeMux(jsonOption)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := pb.RegisterUserServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatalf("server: error while registering gRPC server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// Serve the Swagger UI files using the statik file system.
	statikFS, err := fs.New()
	if err != nil {
		log.Fatalf("server: error while creating statik file system: %v", err)
	}

	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	// Wrap the mux with h2c.NewHandler to support both HTTP/1 and HTTP/2 connections.
	handler := h2c.NewHandler(mux, &http2.Server{})

	// Set the httpServer handler to the wrapped mux.
	server.httpServer.Handler = handler

	// Add a route for the health check service.
	mux.HandleFunc("/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		resp, err := server.healthSrv.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			// Log the error for debugging purposes
			log.Printf("Health check failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
			return
		}

		switch resp.GetStatus() {
		case grpc_health_v1.HealthCheckResponse_SERVING:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("NOT OK"))
		}
	})

	listener, err := net.Listen("tcp", server.config.HttpServerAddress)
	if err != nil {
		log.Fatalf("server: error while creating HTTP listener: %v", err)
	}

	log.Printf("start HTTP Gateway server on %s", listener.Addr().String())
	err = server.httpServer.Serve(listener)
	if err != nil {
		log.Fatalf("server: error while starting HTTP Gateway server: %v", err)
	}
}

func (server *Server) Shutdown() {
	server.grpcServer.GracefulStop()
}

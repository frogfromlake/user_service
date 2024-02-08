package gapi

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

// Server serves gRPC requests for the streamfair user management service.
type Server struct {
	grpcServer *grpc.Server
	pb.UnimplementedUserManagementServiceServer
	config          util.Config
	store           db.Store
	healthSrv       *health.Server
	localTokenMaker token.Maker
}

// NewServer creates a new gRPC server.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	localTokenMaker, err := token.NewLocalPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to create local token maker: %v", err))
	}

	server := &Server{
		config:     config,
		store:      store,
		grpcServer: grpc.NewServer(),
		healthSrv:  health.NewServer(),
		localTokenMaker: localTokenMaker,
	}

	grpc_health_v1.RegisterHealthServer(server.grpcServer, server.healthSrv)

	return server, nil
}

// RunGrpcServer runs a gRPC server on the given address.
func (server *Server) RunGrpcServer() error {
	pb.RegisterUserManagementServiceServer(server.grpcServer, server)
	reflection.Register(server.grpcServer)

	// Set the initial health status to SERVING when the server starts.
	server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	listener, err := net.Listen("tcp", server.config.GrpcServerAddress)
	if err != nil {
		// Update the health status to NOT_SERVING if there's an error.
		server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return fmt.Errorf("server: error while creating listener: %v", err)
	}

	log.Printf("start gRPC server on %s", listener.Addr().String())
	err = server.grpcServer.Serve(listener)
	if err != nil {
		// Update the health status to NOT_SERVING if there's an error.
		server.healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		return fmt.Errorf("server: error while serving gRPC: %v", err)
	}
	return nil
}

// RunGrpcGatewayServer runs a gRPC gateway server on the given address.
func (server *Server) RunGrpcGatewayServer() error {
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

	err := pb.RegisterUserManagementServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		return fmt.Errorf("server: error while registering gRPC server: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	// Add a route for the health check service.
	mux.HandleFunc("/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		resp, err := server.healthSrv.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
		if err != nil {
			// Log the error and respond with an internal server error status code.
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
		return fmt.Errorf("server: error while creating listener: %v", err)
	}

	log.Printf("start HTTP Gateway server on %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		return fmt.Errorf("server: error while starting HTTP Gateway server: %v", err)
	}
	return nil
}

func (server *Server) Shutdown() {
	server.grpcServer.GracefulStop()
}

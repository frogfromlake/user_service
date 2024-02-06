package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/gapi"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
)

func main() {
	fmt.Println("Hello, Streamfair User Service!")
	config, err := util.LoadConfig(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while loading config: %v\n", err)
	}

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: error while parsing config: %v\n", err)
	}

	conn, err := pgxpool.New(context.Background(), poolConfig.ConnString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connection: unable to create connection pool: %v\n", err)
	}

	store := db.NewStore(conn)
	go runGrpcGatewayServer(config, store)
	runGrpcServer(config, store)

}

func runGrpcServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while creating server: %v\n", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, server)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", config.GrpcServerAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while creating listener: %v\n", err)
	}

	log.Printf("start gRPC server on %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while serving gRPC: %v\n", err)
	}
}

func runGrpcGatewayServer(config util.Config, store db.Store) {
	server, err := gapi.NewServer(config, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while creating server: %v\n", err)
	}

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

	err = pb.RegisterUserServiceHandlerServer(ctx, grpcMux, server)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while registering gRPC server: %v\n", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.HttpServerAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while creating listener: %v\n", err)
	}

	log.Printf("start HTTP Gateway server on %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		fmt.Fprintf(os.Stderr, "server: error while starting HTTP Gateway server: %v\n", err)
	}
}

// func runGinServer(config util.Config, store db.Store) {
// 	server, err := api.NewServer(config, store)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "server: error while creating server: %v\n", err)
// 	}

// 	err = server.StartServer(config.HttpServerAddress)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "server: error while starting server: %v\n", err)
// 	}
// }

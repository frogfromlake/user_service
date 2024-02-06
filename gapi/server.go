package gapi

import (
	"fmt"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
)

// Server serves gRPC requests for the streamfair user management service.
type Server struct {
	pb.UnimplementedUserServiceServer
	config          util.Config
	store           db.Store
	localTokenMaker token.Maker
}

// NewServer creates a new gRPC server.
func NewServer(config util.Config, store db.Store) (*Server, error) {
	localTokenMaker, err := token.NewLocalPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		panic(fmt.Sprintf("Failed to create local token maker: %v", err))
	}

	server := &Server{
		config:          config,
		store:           store,
		localTokenMaker: localTokenMaker,
	}

	return server, nil
}

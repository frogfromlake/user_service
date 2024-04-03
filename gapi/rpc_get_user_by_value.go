package gapi

import (
	"context"

	pb "github.com/Streamfair/streamfair_user_svc/common_proto/UserService/pb/user"
	"github.com/Streamfair/streamfair_user_svc/validator"
	"google.golang.org/grpc/codes"
)

func (server *Server) GetUserByValue(ctx context.Context, req *pb.GetUserByValueRequest) (*pb.GetUserByValueResponse, error) {
	username := req.GetUsername()

	// Perform field validation
	err := validator.ValidateUsername(username)
	if err != nil {
		violation := (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("username", err)
		return nil, invalidArgumentError(violation)
	}

	// Fetch user from the database
	user, err := server.store.GetUserByValue(ctx, username)
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}

	rsp := &pb.GetUserByValueResponse{
		User: ConvertUser(user),
	}
	return rsp, nil
}

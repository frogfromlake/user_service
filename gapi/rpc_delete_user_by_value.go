package gapi

import (
	"context"

	pb "github.com/Streamfair/streamfair_user_svc/pb/user"
	"github.com/Streamfair/streamfair_user_svc/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) DeleteUserByValue(ctx context.Context, req *pb.DeleteUserByValueRequest) (*emptypb.Empty, error) {
	usernameParam := req.GetUsername()

	// Perform field validation
	err := validator.ValidateUsername(usernameParam)
	if err != nil {
		violation := (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("username", err)
		return nil, invalidArgumentError(violation)
	}

	// Verify the user exists in the database
	_, err = server.store.GetUserByValue(ctx, usernameParam)
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}

	// Delete the user from the database
	err = server.store.DeleteUserByValue(ctx, usernameParam)
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}

	return &emptypb.Empty{}, nil
}

package gapi

import (
	"context"

	"github.com/Streamfair/streamfair_user_svc/validator"
	pb "github.com/Streamfair/streamfair_user_svc/_common_proto/UserService/pb/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (server *Server) DeleteUserById(ctx context.Context, req *pb.DeleteUserByIdRequest) (*emptypb.Empty, error) {
	idParam := req.GetId()
	err := validator.ValidateId(idParam)
	// Perform field validation
	if err != nil {
		violation := (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("id", err)
		return nil, invalidArgumentError(violation)
	}

	// Verify the user exists in the database
	_, err = server.store.GetUserById(ctx, idParam)
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}

	// Delete the user from the database
	err = server.store.DeleteUserById(ctx, idParam)
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}

	return &emptypb.Empty{}, nil
}

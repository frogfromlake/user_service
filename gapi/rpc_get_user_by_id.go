package gapi

import (
	"context"

	pb "github.com/Streamfair/streamfair_user_svc/_common_proto/UserService/pb/user"
	"github.com/Streamfair/streamfair_user_svc/validator"
	"google.golang.org/grpc/codes"
)

func (server *Server) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.GetUserByIdResponse, error) {
	idParam := req.GetId()

	// Perform field validation
	err := validator.ValidateId(idParam)
	if err != nil {
		violation := (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("id", err)
		return nil, invalidArgumentError(violation)
	}

	// Fetch user from the database
	user, err := server.store.GetUserById(ctx, idParam)
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}
	
	rsp := &pb.GetUserByIdResponse{
		User: ConvertUser(user),
	}
	return rsp, nil
}

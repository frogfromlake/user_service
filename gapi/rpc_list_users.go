package gapi

import (
	"context"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	pb "github.com/Streamfair/streamfair_user_svc/common_proto/UserService/pb/user"
	"github.com/Streamfair/streamfair_user_svc/validator"
	"google.golang.org/grpc/codes"
)

func (server *Server) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	// Validate the request
	violations := validateListUsersRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentErrors(violations)
	}

	users, err := server.store.ListUsers(ctx, db.ListUsersParams{
		Limit:  req.GetLimit(),
		Offset: req.GetOffset(),
	})
	if err != nil {
		// Handle database errors
		return nil, handleDatabaseError(err)
	}

	rsp := &pb.ListUsersResponse{
		Users: convertUsersList(users),
	}

	return rsp, nil
}

func validateListUsersRequest(req *pb.ListUsersRequest) (violations []*CustomError) {
	if err := validator.ValidateLimit(req.GetLimit()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.OutOfRange,
		}).WithDetails("limit", err))
	}

	if err := validator.ValidateOffset(req.GetOffset()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.OutOfRange,
		}).WithDetails("offset", err))
	}

	return violations
}

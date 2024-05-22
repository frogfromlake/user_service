package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	pb "github.com/Streamfair/common_proto/UserService/pb/user"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/Streamfair/streamfair_user_svc/validator"
	"google.golang.org/grpc/codes"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	violations := validateCreateUserRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentErrors(violations)
	}

	arg := db.CreateUserParams{
		Username:     req.GetUsername(),
		FullName:     req.GetFullName(),
		Email:        req.GetEmail(),
		PasswordHash: req.GetPasswordHash(),
		PasswordSalt: req.GetPasswordSalt(),
		CountryCode:  req.GetCountryCode(),
		RoleID:       util.ConvertToInt8(req.GetRoleId()),
		Status:       util.ConvertToText(req.GetStatus()),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		log.Printf("failed to create user: %v", err)
		if strings.Contains(err.Error(), "Users_email_key") {
			violation := (&CustomError{
				StatusCode: codes.AlreadyExists,
			}).WithDetails("email", fmt.Errorf("user with email %s already exists", req.GetEmail()))
			return nil, invalidArgumentError(violation)
		} else if strings.Contains(err.Error(), "Users_username_key") {
			violation := (&CustomError{
				StatusCode: codes.AlreadyExists,
			}).WithDetails("username", fmt.Errorf("user with username %s already exists", req.GetUsername()))
			return nil, invalidArgumentError(violation)
		}
		return nil, handleDatabaseError(err)
	}

	rsp := &pb.CreateUserResponse{
		User: ConvertUser(user),
	}

	return rsp, nil
}

// validateCreateTokenRequest validates the create token request and returns a slice of custom errors.
func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*CustomError) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("username", err))
	}

	if err := validator.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("full_name", err))
	}

	if err := validator.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("email", err))
	}

	if err := validator.ValidateCountryCode(req.GetCountryCode()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("country_code", err))
	}

	if err := validator.ValidateRoleId(req.GetRoleId()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("role_id", err))
	}

	if err := validator.ValidateStatus(req.GetStatus()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("status", err))
	}

	return violations
}

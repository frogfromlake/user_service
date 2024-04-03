package gapi

import (
	"context"
	"time"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	pb "github.com/Streamfair/streamfair_user_svc/_common_proto/UserService/pb/user"
	"github.com/Streamfair/streamfair_user_svc/validator"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc/codes"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	// Validate the request
	violations := validateUpdateUserRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentErrors(violations)
	}

	user, err := server.store.GetUserById(ctx, req.GetId())
	if err != nil {
		return nil, handleDatabaseError(err)
	}

	arg := db.UpdateUserParams{
		Username:          pgtype.Text{String: req.GetUsername(), Valid: req.Username != "" && req.Username != user.Username},
		UsernameChangedAt: pgtype.Timestamptz{Time: time.Now(), Valid: req.UsernameChangedAt != nil},
		FullName:          pgtype.Text{String: req.GetFullName(), Valid: req.FullName != ""},
		Email:             pgtype.Text{String: req.GetEmail(), Valid: req.Email != "" && req.Email != user.Email},
		EmailChangedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: req.EmailChangedAt != nil},
		PasswordHash:      pgtype.Text{String: req.GetPasswordHash(), Valid: req.PasswordHash != ""},
		PasswordSalt:      pgtype.Text{String: req.GetPasswordSalt(), Valid: req.PasswordSalt != ""},
		PasswordChangedAt: pgtype.Timestamptz{Time: time.Now(), Valid: req.PasswordChangedAt != nil},
		CountryCode:       pgtype.Text{String: req.GetCountryCode(), Valid: req.CountryCode != ""},
		RoleID:            pgtype.Int8{Int64: req.GetRoleId(), Valid: req.RoleId != 0},
		Status:            pgtype.Text{String: req.GetStatus(), Valid: req.Status != ""},
		ID:                req.GetId(),
	}

	// Update the user in the database
	user, err = server.store.UpdateUser(ctx, arg)
	if err != nil {
		return nil, handleDatabaseError(err)
	}

	rsp := &pb.UpdateUserResponse{
		User: ConvertUser(user),
	}

	return rsp, nil
}

// validateCreateTokenRequest validates the create token request and returns a slice of custom errors.
func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*CustomError) {
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

	if err := validator.ValidateRoleId(req.GetId()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("id", err))
	}

	if err := validator.ValidateStatus(req.GetStatus()); err != nil {
		violations = append(violations, (&CustomError{
			StatusCode: codes.InvalidArgument,
		}).WithDetails("status", err))
	}

	return violations
}

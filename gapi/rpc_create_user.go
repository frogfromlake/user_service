package gapi

import (
	"context"
	"encoding/base64"
	"errors"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	pb "github.com/Streamfair/streamfair_user_svc/pb/user"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {

	byteHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	hashedPassword := base64.StdEncoding.EncodeToString(byteHash.Hash)
	passwordSalt := base64.StdEncoding.EncodeToString(byteHash.Salt)

	arg := db.CreateUserParams{
		Username:     req.GetUsername(),
		FullName:     req.GetFullName(),
		Email:        req.GetEmail(),
		PasswordHash: hashedPassword,
		PasswordSalt: passwordSalt,
		CountryCode:  req.GetCountryCode(),
		RoleID:       util.ConvertToInt8(req.GetRoleId()),
		Status:       util.ConvertToText(req.GetStatus()),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				return nil, status.Errorf(codes.AlreadyExists, "unique violation occured: %v", err)
			case "23503": // foreign_key_violation
				return nil, status.Errorf(codes.FailedPrecondition, "foreign key violation occurred: %v", err)
			default:
				return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	rsp := &pb.CreateUserResponse{
		User: convertUser(user),
	}

	return rsp, nil
}

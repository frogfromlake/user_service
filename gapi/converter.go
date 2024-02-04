package gapi

import (
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func convertUser(user db.UserSvcUser) *pb.User {
	return &pb.User{
		Id:                user.ID,
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CountryCode:       user.CountryCode,
		RoleId:            user.RoleID.Int64,
		Status:            user.Status.String,
		LastLoginAt:       timestamppb.New(user.LastLoginAt),
		UsernameChangedAt: timestamppb.New(user.UsernameChangedAt),
		EmailChangedAt:    timestamppb.New(user.EmailChangedAt),
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
	}
}

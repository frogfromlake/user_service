package gapi

import (
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	pb "github.com/Streamfair/streamfair_user_svc/_common_proto/UserService/pb/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertUser(user db.UserSvcUser) *pb.User {
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

func convertUsersList(users []db.ListUsersRow) []*pb.Users {
	// TODO: Check if i have to use User or Users
	var userList []*pb.Users
	for _, user := range users {
		userList = append(userList, &pb.Users{
			Id:                user.ID,
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			CountryCode:       user.CountryCode,
			RoleId:            user.RoleID.Int64,
			LastLoginAt:       timestamppb.New(user.LastLoginAt),
			CreatedAt:         timestamppb.New(user.CreatedAt),
			UpdatedAt:         timestamppb.New(user.UpdatedAt),
		})
	}
	return userList
}
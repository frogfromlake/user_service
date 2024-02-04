package gapi

import (
	"context"
	"database/sql"
	"encoding/base64"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/pb"
	"github.com/Streamfair/streamfair_user_svc/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) LoginUser(ctx context.Context, req *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {

	user, err := server.store.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "user not found: %v", err)
	}

	byteHash, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "decoding error occured.")
	}
	byteSalt, err := base64.StdEncoding.DecodeString(user.PasswordSalt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "decoding error occured.")
	}

	err = util.ComparePassword(byteHash, byteSalt, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "incorrect password.")
	}

	accessToken, accessPayload, err := server.localTokenMaker.CreateLocalToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token.")
	}

	refreshToken, refreshPayload, err := server.localTokenMaker.CreateLocalToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token.")
	}

	mtdt := server.extractMetadata(ctx)
	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent, // TODO
		ClientIp:     mtdt.ClientIP, // TODO
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v.", err)
	}

	rps := &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}

	return rps, nil
}

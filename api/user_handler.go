package api

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type createUserRequest struct {
	Username    string `json:"username" binding:"required,min=3"`
	FullName    string `json:"full_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=64"`
	CountryCode string `json:"country_code" binding:"required,iso3166_1_alpha2"`
	RoleID      int64  `json:"role_id" binding:"required,min=1,max=3"`
	Status      string `json:"status" binding:"required,oneof=active inactive"`
}
type userResponse struct {
	ID                int64     `json:"id"`
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	CountryCode       string    `json:"country_code"`
	RoleID            int64     `json:"role_id"`
	Status            string    `json:"status"`
	LastLoginAt       time.Time `json:"last_login_at"`
	UsernameChangedAt time.Time `json:"username_changed_at"`
	EmailChangedAt    time.Time `json:"email_changed_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func newUserResponse(user db.UserSvcUser) userResponse {
	return userResponse{
		ID:                user.ID,
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CountryCode:       user.CountryCode,
		RoleID:            user.RoleID.Int64,
		Status:            user.Status.String,
		LastLoginAt:       user.LastLoginAt,
		UsernameChangedAt: user.UsernameChangedAt,
		EmailChangedAt:    user.EmailChangedAt,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	byteHash, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	hashedPassword := base64.StdEncoding.EncodeToString(byteHash.Hash)
	passwordSalt := base64.StdEncoding.EncodeToString(byteHash.Salt)

	arg := db.CreateUserParams{
		Username:     req.Username,
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		PasswordSalt: passwordSalt,
		CountryCode:  req.CountryCode,
		RoleID:       util.ConvertToInt8(req.RoleID),
		Status:       util.ConvertToText(req.Status),
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // unique_violation
				ctx.JSON(http.StatusConflict, errorResponse(err))
			case "23503": // foreign_key_violation
				ctx.JSON(http.StatusConflict, errorResponse(err))
			default:
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			}
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(user)
	ctx.JSON(http.StatusOK, rsp)
}

type getUserByIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}
type getUserByIDResponse struct {
	ID                int64     `json:"id"`
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	CountryCode       string    `json:"country_code"`
	RoleID            int64     `json:"role_id" binding:"required,min=1,max=3"`
	Status            string    `json:"status" binding:"required,oneof=active inactive"`
	LastLoginAt       time.Time `json:"last_login_at"`
	UsernameChangedAt time.Time `json:"username_changed_at"`
	EmailChangedAt    time.Time `json:"email_changed_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (server *Server) getUserByID(ctx *gin.Context) {
	var req getUserByIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByID(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	rsp := getUserByIDResponse{
		ID:                user.ID,
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CountryCode:       user.CountryCode,
		RoleID:            user.RoleID.Int64,
		Status:            user.Status.String,
		LastLoginAt:       user.LastLoginAt,
		UsernameChangedAt: user.UsernameChangedAt,
		EmailChangedAt:    user.EmailChangedAt,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}

type getUserByUsernameRequest struct {
	Username string `uri:"username" binding:"required,min=3"`
}
type getUserByUsernameResponse struct {
	ID                int64     `json:"id"`
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	CountryCode       string    `json:"country_code"`
	RoleID            int64     `json:"role_id"`
	Status            string    `json:"status"`
	LastLoginAt       time.Time `json:"last_login_at"`
	UsernameChangedAt time.Time `json:"username_changed_at"`
	EmailChangedAt    time.Time `json:"email_changed_at"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (server *Server) getUserByUsername(ctx *gin.Context) {
	var req getUserByUsernameRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByValue(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	rsp := getUserByUsernameResponse{
		ID:                user.ID,
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		CountryCode:       user.CountryCode,
		RoleID:            user.RoleID.Int64,
		Status:            user.Status.String,
		LastLoginAt:       user.LastLoginAt,
		UsernameChangedAt: user.UsernameChangedAt,
		EmailChangedAt:    user.EmailChangedAt,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, rsp)
}

type listUsersRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

func (server *Server) listUsers(ctx *gin.Context) {
	var req listUsersRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListUsersParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	users, err := server.store.ListUsers(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, users)
}

type updateUserUri struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}
type updateUserRequest struct {
	Username          pgtype.Text        `json:"username" binding:"omitempty,min=3"`
	UsernameChangedAt pgtype.Timestamptz `json:"username_changed_at" binding:"omitempty"`
	FullName          pgtype.Text        `json:"full_name" binding:"omitempty"`
	Email             pgtype.Text        `json:"email" binding:"omitempty,email"`
	EmailChangedAt    pgtype.Timestamptz `json:"email_changed_at" binding:"omitempty"`
	Password          pgtype.Text        `json:"password" binding:"omitempty,min=8,max=64"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at" binding:"omitempty"`
	CountryCode       pgtype.Text        `json:"country_code" binding:"omitempty,iso3166_1_alpha2"`
	RoleID            pgtype.Int8        `json:"role_id" binding:"omitempty,min=1,max=3"`
	Status            pgtype.Text        `json:"status" binding:"omitempty,oneof=active inactive"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var uri updateUserUri
	var req updateUserRequest
	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	byteHash, err := util.HashPassword(req.Password.String)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	hashedPassword := base64.StdEncoding.EncodeToString(byteHash.Hash)
	passwordSalt := base64.StdEncoding.EncodeToString(byteHash.Salt)

	arg := db.UpdateUserParams{
		Username:          req.Username,
		UsernameChangedAt: pgtype.Timestamptz{Time: time.Now()},
		FullName:          req.FullName,
		Email:             req.Email,
		EmailChangedAt:    pgtype.Timestamptz{Time: time.Now()},
		PasswordHash:      pgtype.Text{String: hashedPassword},
		PasswordSalt:      pgtype.Text{String: passwordSalt},
		PasswordChangedAt: pgtype.Timestamptz{Time: time.Now()},
		CountryCode:       req.CountryCode,
		RoleID:            req.RoleID,
		Status:            req.Status,
	}

	user, err := server.store.UpdateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type deleteUserRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := server.store.GetUserByID(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	err = server.store.DeleteUser(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "user deleted successfully!"})
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.store.GetUserByValue(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	byteHash, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	byteSalt, err := base64.StdEncoding.DecodeString(user.PasswordSalt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = util.ComparePassword(byteHash, byteSalt, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessPayload, err := server.localTokenMaker.CreateLocalToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshPayload, err := server.localTokenMaker.CreateLocalToken(
		user.Username,
		server.config.RefreshTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, rsp)
}

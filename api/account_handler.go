package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/Streamfair/user_service/db/sqlc"
	"github.com/Streamfair/user_service/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
Goal:

DONE: When a new user creates an account, he should have the option to select additionnal accounttypes for his
account (artist, label, producer, writer). He should be able to choose all of them or pick the ones he need.

TODO: Also after account creation the user should be able to:
1.) switch to those created types (for example switch to his artist account to upload a song).
2.) add or delete specific accounttypes from his account.

TODO:
1. Switching to different account types:
2. Adding or deleting account types:
*/
type createAccountRequest struct {
	AccountTypeID []int64     `json:"account_type_id" binding:"required,min=1,acctype"`
	Username      string      `json:"username" binding:"required,min=3"`
	Email         string      `json:"email" binding:"required,email"`
	PasswordHash  string      `json:"password_hash" binding:"required"` // tag: sha256
	CountryCode   string      `json:"country_code" binding:"required,iso3166_1_alpha2"`
	AvatarUrl     pgtype.Text `json:"avatar_url"` // rename to uri, tag: uri
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountTxParams{
		AccountParams: db.CreateAccountParams{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: req.PasswordHash,
			CountryCode:  req.CountryCode,
			AvatarUrl:    req.AvatarUrl,
		},
		AccountTypeID: req.AccountTypeID,
	}

	account, err := server.store.CreateAccountTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountByIDRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type getAccountByIDResponse struct {
	Account        db.GetAccountByIDRow
	AccountTypeIDs []db.GetAccountTypeIDsForAccountRow
}

func (server *Server) getAccountByID(ctx *gin.Context) {
	var req getAccountByIDRequest
	var res getAccountByIDResponse
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccountByID(ctx, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.Account = account

	accountTypeIDs, err := server.store.GetAccountTypeIDsForAccount(ctx, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.AccountTypeIDs = accountTypeIDs

	ctx.JSON(http.StatusOK, res)
}

func (server *Server) handleMissingUsername(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("missing username")))
}

type getAccountByUsernameRequest struct {
	Username string `uri:"username" binding:"required,min=3"`
}

type getAccountByUsernameResponse struct {
	Account        db.GetAccountByUsernameRow
	AccountTypeIDs []db.GetAccountTypeIDsForAccountRow
}

func (server *Server) getAccountByUsername(ctx *gin.Context) {
	var req getAccountByUsernameRequest
	var res getAccountByUsernameResponse
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccountByUsername(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.Account = account

	accountTypeIDs, err := server.store.GetAccountTypeIDsForAccount(ctx, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.AccountTypeIDs = accountTypeIDs

	ctx.JSON(http.StatusOK, res)
}

type getAccountByAllParamsRequest struct {
	Username    string `form:"username" binding:"required,min=3"`
	Email       string `form:"email" binding:"required,email"`
	CountryCode string `form:"country_code" binding:"required,iso3166_1_alpha2"`
	AvatarUrl   string `form:"avatar_url" binding:"required"`
}

type getAccountByAllParamsResponse struct {
	Account        db.UserServiceAccount
	AccountTypeIDs []db.GetAccountTypeIDsForAccountRow
}

func (server *Server) getAccountbyAllParams(ctx *gin.Context) {
	var req getAccountByAllParamsRequest
	var res getAccountByAllParamsResponse
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetAccountByAllParamsParams{
		Username:    req.Username,
		Email:       req.Email,
		CountryCode: req.CountryCode,
		AvatarUrl:   util.ConvertToText(req.AvatarUrl),
	}

	account, err := server.store.GetAccountByAllParams(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.Account = account

	accountTypeIDs, err := server.store.GetAccountTypeIDsForAccount(ctx, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.AccountTypeIDs = accountTypeIDs

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=100"`
}

func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

// TODO: should be able to update account type
type updateAccountURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateAccountRequest struct {
	Username     string      `json:"username" binding:"omitempty,min=3"`
	Email        string      `json:"email" binding:"omitempty,email"`
	CountryCode  string      `json:"country_code" binding:"omitempty,iso3166_1_alpha2"`
	AvatarUrl    pgtype.Text `json:"avatar_url" binding:"omitempty"`
	LikesCount   int64       `json:"likes_count" binding:"omitempty"`
	FollowsCount int64       `json:"follows_count" binding:"omitempty"`
}

func (server *Server) updateAccount(ctx *gin.Context) {
	var uri updateAccountURI
	var req updateAccountRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateAccountParams{
		ID:           uri.ID,
		Username:     req.Username,
		Email:        req.Email,
		CountryCode:  req.CountryCode,
		AvatarUrl:    req.AvatarUrl,
		LikesCount:   req.LikesCount,
		FollowsCount: req.FollowsCount,
	}

	account, err := server.store.UpdateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type updateAccountPasswordURI struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateAccountPasswordRequest struct {
	PasswordHash string `json:"password_hash" binding:"required,min=8"`
}

func (server *Server) updateAccountPassword(ctx *gin.Context) {
	var uri updateAccountPasswordURI
	var req updateAccountPasswordRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.UpdateAccountPasswordParams{
		ID:           uri.ID,
		PasswordHash: req.PasswordHash,
	}

	account, err := server.store.UpdateAccountPassword(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type deleteAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAccount(ctx *gin.Context) {
	var req deleteAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	_, err := server.store.GetAccountByID(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	err = server.store.DeleteAccountTx(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "account deleted successfully!"})
}

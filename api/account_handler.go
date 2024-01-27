package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
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
	AccountTypeIDs []int64     `json:"account_type_ids" binding:"required,min=1,acctype"`
	Owner          string      `json:"owner" binding:"required,min=3"`
	AvatarUri      pgtype.Text `json:"avatar_uri"` // tag: uri
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountTxParams{
		AccountTypeIDs: req.AccountTypeIDs,
		AccountParams: db.CreateAccountParams{
			Owner:     req.Owner,
			AvatarUrl: req.AvatarUri,
		},
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
	Account      db.UserSvcAccount
	AccountTypes []db.UserSvcAccountType
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

	accountTypes, err := server.store.GetAccountTypesForAccount(ctx, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.AccountTypes = accountTypes

	ctx.JSON(http.StatusOK, res)
}

func (server *Server) handleMissingUsername(ctx *gin.Context) {
	ctx.JSON(http.StatusBadRequest, errorResponse(fmt.Errorf("missing owner in request")))
}

type getAccountByOwnerRequest struct {
	Owner string `uri:"owner" binding:"required,min=3"`
}

type getAccountByOwnerResponse struct {
	Account      db.UserSvcAccount
	AccountTypes []db.UserSvcAccountType
}

func (server *Server) getAccountByOwner(ctx *gin.Context) {
	var req getAccountByOwnerRequest
	var res getAccountByOwnerResponse
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccountByOwner(ctx, req.Owner)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.Account = account

	accountTypes, err := server.store.GetAccountTypesForAccount(ctx, account.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res.AccountTypes = accountTypes

	ctx.JSON(http.StatusOK, res)
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
	Owner     string      `json:"username" binding:"omitempty,min=3"`
	AvatarUrl pgtype.Text `json:"avatar_url" binding:"omitempty"`
	Plays     int64       `json:"plays" binding:"omitempty"`
	Likes     int64       `json:"likes" binding:"omitempty"`
	Follows   int64       `json:"follows" binding:"omitempty"`
	Shares    int64       `json:"shares" binding:"omitempty"`
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
		ID:        uri.ID,
		Owner:     req.Owner,
		AvatarUrl: req.AvatarUrl,
		Plays:     req.Plays,
		Likes:     req.Likes,
		Follows:   req.Follows,
		Shares:    req.Shares,
	}

	account, err := server.store.UpdateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// type updateAccountPasswordURI struct {
// 	ID int64 `uri:"id" binding:"required,min=1"`
// }

// type updateAccountPasswordRequest struct {
// 	PasswordHash string `json:"password_hash" binding:"required,min=8"`
// }

// func (server *Server) updateAccountPassword(ctx *gin.Context) {
// 	var uri updateAccountPasswordURI
// 	var req updateAccountPasswordRequest

// 	if err := ctx.ShouldBindUri(&uri); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	arg := db.UpdateAccountPasswordParams{
// 		ID: uri.ID,
// 	}

// 	account, err := server.store.UpdateAccountPassword(ctx, arg)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, account)
// }

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

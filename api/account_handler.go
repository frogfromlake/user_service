package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/token"
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
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
	AccountType int32  `json:"account_type" binding:"required,min=1,acctype"`
	AvatarUri   string `json:"avatar_uri" binding:"uri"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountTxParams{
		AccountParams: db.CreateAccountParams{
			Owner:       authPayload.Username,
			AccountType: req.AccountType,
			AvatarUri:   util.ConvertToText(req.AvatarUri),
		},
	}

	account, err := server.store.CreateAccountTx(ctx, arg)
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

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, res)
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

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

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

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner: authPayload.Username,
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
type updateAccountUri struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

type updateAccountRequest struct {
	AvatarUri pgtype.Text `json:"avatar_url" binding:"omitempty"`
	Plays     int64       `json:"plays" binding:"omitempty"`
	Likes     int64       `json:"likes" binding:"omitempty"`
	Follows   int64       `json:"follows" binding:"omitempty"`
	Shares    int64       `json:"shares" binding:"omitempty"`
}

func (server *Server) updateAccount(ctx *gin.Context) {
	var uri updateAccountUri
	var req updateAccountRequest

	if err := ctx.ShouldBindUri(&uri); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.UpdateAccountParams{
		ID:        uri.ID,
		Owner:     authPayload.Username,
		AvatarUri: req.AvatarUri,
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

type deleteAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAccount(ctx *gin.Context) {
	var req deleteAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccountByID(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	err = server.store.DeleteAccountTx(ctx, req.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "account deleted successfully!"})
}

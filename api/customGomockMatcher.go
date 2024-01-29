package api

import (
	"encoding/base64"
	"fmt"
	"reflect"

	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"go.uber.org/mock/gomock"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x any) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	hash, _ := base64.StdEncoding.DecodeString(arg.PasswordHash)
	salt, _ := base64.StdEncoding.DecodeString(arg.PasswordSalt)

	err := util.ComparePassword(hash, salt, e.password)
	if err != nil {
		return false
	}

	e.arg.PasswordHash = arg.PasswordHash
	e.arg.PasswordSalt = arg.PasswordSalt
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg: %v and password: (%v)", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

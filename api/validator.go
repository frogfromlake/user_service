package api

import (
	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/go-playground/validator/v10"
)

var validAccountTypes validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if accountTypes, ok := fieldLevel.Field().Interface().(int32); ok {
		// check if account type is supported.
		return isSupportedAccountType(accountTypes)
	}
	return false
}

// IsSupportedAccountType is a helper for an endpoint validator
// to check if the account type is supported. It uses accountTypes slice to check.
func isSupportedAccountType(accountType int32) bool {
	supportedTypes := util.GetAccountTypeStruct()
	for _, supported := range supportedTypes {
		if accountType == supported.ID {
			return true
		}
	}
	return false
}
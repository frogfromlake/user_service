package api

import (
	"github.com/frogfromlake/streamfair_backend/user_service/util"
	"github.com/go-playground/validator/v10"
)

var validAccountTypes validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if accountTypes, ok := fieldLevel.Field().Interface().([]int64); ok {
		// check if account type is supported.
		return isSupportedAccountType(accountTypes)
	}
	return false
}

// IsSupportedAccountType is a helper for an endpoint validator
// to check if the account type is supported. It uses accountTypes slice to check.
func isSupportedAccountType(accountType []int64) bool {
	supportedTypes := util.GetAccountTypeStruct()
	for _, v := range accountType {
		found := false
		for _, supported := range supportedTypes {
			if v == supported.ID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

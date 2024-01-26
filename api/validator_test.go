package api

import (
	"testing"

	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/stretchr/testify/require"
)

func TestIsSupportedAccountType(t *testing.T) {
	// Define a slice of supported account types
	supportedTypes := util.GetAccountTypeStruct()

	// Convert supportedTypes to []int64
	var supportedTypeIDs []int64
	for _, typ := range supportedTypes {
		supportedTypeIDs = append(supportedTypeIDs, typ.ID)
	}

	// Test with a supported account type
	require.True(t, isSupportedAccountType(supportedTypeIDs))
	require.True(t, isSupportedAccountType([]int64{1}))

	// Test with an unsupported account type
	require.False(t, isSupportedAccountType([]int64{99999}))

	// Test with a mix of supported and unsupported account types
	require.False(t, isSupportedAccountType([]int64{1, 99999}))
}

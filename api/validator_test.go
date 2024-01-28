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
	var supportedTypeIDs []int32
	for _, typ := range supportedTypes {
		supportedTypeIDs = append(supportedTypeIDs, typ.ID)
	}

	// Test with a supported account type
	require.True(t, isSupportedAccountType(supportedTypeIDs[0]))

	// Test with an unsupported account type
	require.False(t, isSupportedAccountType(99999))

	// Test with a mix of supported and unsupported account types
	// Since we're testing a single account type now, let's just pick the first supported type
	require.True(t, isSupportedAccountType(supportedTypeIDs[0]))
}

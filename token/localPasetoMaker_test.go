package token

import (
	"testing"
	"time"

	"github.com/Streamfair/streamfair_user_svc/util"
	"github.com/stretchr/testify/require"
)

func TestLocalPasetoMaker(t *testing.T) {
	maker, err := NewLocalPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomUsername()
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, err := maker.CreateLocalToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyLocalToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
}

func TestExpiredToken(t *testing.T) {
	maker, err := NewLocalPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomUsername()
	duration := -time.Minute

	token, err := maker.CreateLocalToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyLocalToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidToken(t *testing.T) {
	maker, err := NewLocalPasetoMaker(util.RandomString(32))
	require.NoError(t, err)

	payload, err := maker.VerifyLocalToken("invalid_token")
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}

package api

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sort"
	"testing"
	"time"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
	db "github.com/Streamfair/streamfair_user_svc/db/sqlc"
	"github.com/Streamfair/streamfair_user_svc/util"
	"go.uber.org/mock/gomock"
)

func TestStartServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_db.NewMockStore(ctrl)

	testCases := []struct {
		name        string
		pingError   bool
		invalidPort bool
		dbError     bool
	}{
		{
			name:        "ServerStartSuccessful",
			pingError:   false,
			invalidPort: false,
		},
		{
			name:        "ServerStartFailureInvalidPort",
			pingError:   false,
			invalidPort: true,
		},
		{
			name:        "ServerStartFailurePingError",
			pingError:   true,
			invalidPort: false,
		},
		{
			name:        "InitializeDatabaseError",
			pingError:   false,
			invalidPort: false,
			dbError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "ServerStartSuccessful" || tc.pingError {
				mockStore.EXPECT().Ping(gomock.Any(), gomock.Any()).Times(1).Return(nil)
			}
			if tc.dbError {
				mockStore.EXPECT().ListAccountTypes(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("some error"))
			} else {
				mockStore.EXPECT().ListAccountTypes(gomock.Any(), gomock.Any()).Times(1).Return([]db.UserServiceAccountType{{ID: 1}, {ID: 2}}, nil)
				mockStore.EXPECT().CreateAccountType(gomock.Any(), gomock.Any()).AnyTimes().Return(db.UserServiceAccountType{}, nil)
			}

			server := NewServer(mockStore)
			port, err := getRandomPort()
			if tc.invalidPort {
				port = -1
			}
			if (err != nil || port <= 0) && !tc.invalidPort {
				t.Fatalf("Failed to get random port: %v", err)
			}

			go func() {
				err := server.StartServer(fmt.Sprintf("localhost:%d", port))
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			}()

			time.Sleep(500 * time.Millisecond)
			startTime := time.Now()
			if !tc.invalidPort && !tc.dbError {
				for {
					resp, err := http.Get(fmt.Sprintf("http://localhost:%d/readiness", port))
					if err == nil && resp.StatusCode == http.StatusOK {
						break
					}
					if time.Since(startTime) > 5*time.Second {
						t.Errorf("failed to send GET request within 5 seconds")
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		})
	}
}

func getRandomPort() (int, error) {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr := ln.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func convertUtilAccountTypesToDbAccountTypes(utilAccountTypes []util.AccountType) []db.UserServiceAccountType {
	dbAccountTypes := make([]db.UserServiceAccountType, len(utilAccountTypes))
	for i, utilAccountType := range utilAccountTypes {
		dbAccountTypes[i] = db.UserServiceAccountType{
			ID:          utilAccountType.ID,
			Description: utilAccountType.Description,
			Permissions: utilAccountType.Permissions,
			IsArtist:    utilAccountType.IsArtist,
			IsProducer:  utilAccountType.IsProducer,
			IsWriter:    utilAccountType.IsWriter,
			IsLabel:     utilAccountType.IsLabel,
		}
	}
	return dbAccountTypes
}

func TestInitializeDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStore := mock_db.NewMockStore(ctrl)

	utilAccountTypes := util.GetAccountTypeStruct()
	dbAccountTypes := convertUtilAccountTypesToDbAccountTypes(utilAccountTypes)

	testCases := []struct {
		name         string
		accountTypes []db.UserServiceAccountType
		buildStubs   func(*mock_db.MockStore)
		expectedErr  error
		expectedStr  []string
	}{
		{
			name:         "OK",
			accountTypes: dbAccountTypes,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().ListAccountTypes(gomock.Any(), gomock.Any()).Times(1).Return([]db.UserServiceAccountType{}, nil)
				for _, accountType := range dbAccountTypes {
					store.EXPECT().CreateAccountType(gomock.Any(), db.CreateAccountTypeParams{
						Description: accountType.Description,
						Permissions: accountType.Permissions,
						IsArtist:    accountType.IsArtist,
						IsProducer:  accountType.IsProducer,
						IsWriter:    accountType.IsWriter,
						IsLabel:     accountType.IsLabel,
					}).Times(1).Return(accountType, nil)
				}
			},
		},
		{
			name:         "ListAccountTypesError",
			accountTypes: dbAccountTypes,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().ListAccountTypes(gomock.Any(), gomock.Any()).Times(1).Return(nil, errors.New("some error"))
			},
			expectedErr: errors.New("some error"),
		},
		{
			name:         "CreateAccountTypeError",
			accountTypes: dbAccountTypes,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().ListAccountTypes(gomock.Any(), gomock.Any()).Times(1).Return([]db.UserServiceAccountType{}, nil)
				for _, accountType := range dbAccountTypes {
					store.EXPECT().CreateAccountType(gomock.Any(), db.CreateAccountTypeParams{
						Description: accountType.Description,
						Permissions: accountType.Permissions,
						IsArtist:    accountType.IsArtist,
						IsProducer:  accountType.IsProducer,
						IsWriter:    accountType.IsWriter,
						IsLabel:     accountType.IsLabel,
					}).Times(1).Return(db.UserServiceAccountType{}, errors.New("another error"))
				}
			},
			expectedStr: []string{"another error", "another error", "another error", "another error", "another error"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(mockStore)

			err := InitializeDatabase(mockStore)

			// Check if expectedErrs is not empty
			if len(tc.expectedStr) > 0 {
				// Sort both slices
				sort.Strings(tc.expectedStr)

				// Convert err to a slice of strings for comparison
				errStrs := append([]string(nil), tc.expectedStr...)

				// Compare both sorted slices
				if !reflect.DeepEqual(tc.expectedStr, errStrs) {
					t.Errorf("Expected errors %v, got %v", tc.expectedStr, errStrs)
				}
			} else if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("Expected error, got none")
				} else if err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error %q, got %q", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error, got %q", err)
				}
			}
		})
	}
}

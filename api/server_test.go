package api

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	mock_db "github.com/Streamfair/streamfair_user_svc/db/mock"
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

			server := newTestServer(t, mockStore)
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

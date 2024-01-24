package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	mock_db "github.com/frogfromlake/streamfair_backend/user_service/db/mock"
	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func TestReadinessCheck(t *testing.T) {
	tests := []struct {
		name      string
		pingError bool
		expected  int
	}{
		{
			name:      "Database is ready",
			pingError: false,
			expected:  http.StatusOK,
		},
		{
			name:      "Database is not ready",
			pingError: true,
			expected:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_db.NewMockStore(ctrl)
			server := &Server{store: mockStore}
			router := gin.Default()
			router.GET("/readiness", server.readinessCheck)

			if tt.pingError {
				mockStore.EXPECT().Ping(gomock.Any(), gomock.Any()).Return(errors.New("database not ready"))
			} else {
				mockStore.EXPECT().Ping(gomock.Any(), gomock.Any()).Return(nil)
			}

			req, _ := http.NewRequest("GET", "/readiness", nil)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != tt.expected {
				t.Errorf("Expected status code %d, got %d", tt.expected, resp.Code)
			}

			if tt.pingError && !strings.Contains(resp.Body.String(), "Database not ready") {
				t.Errorf("Expected 'Database not ready' message, got %s", resp.Body.String())
			}
		})
	}
}

package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/mocks"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestLoginHandler(t *testing.T) {
	mockStorage := new(mocks.Storage) // Используем новый мок
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
		Secret:  "test_secret",
	}

	cases := []struct {
		name           string
		requestBody    models.User
		mockResponse   *models.UserEntry
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			requestBody:    models.User{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "User not found",
			requestBody:    models.User{Login: "nonexistent", Password: "password"},
			mockError:      models.ErrUserNotFound,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Wrong password",
			requestBody:    models.User{Login: "user", Password: "wrongpass"},
			mockResponse:   &models.UserEntry{ID: 1, Username: "user", PasswordHash: "correctpass", Balance: 0.0},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Successful login",
			requestBody:    models.User{Login: "user", Password: "correctpass"},
			mockResponse:   &models.UserEntry{ID: 1, Username: "user", PasswordHash: "correctpass", Balance: 0.0},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockResponse != nil && tt.mockResponse.PasswordHash != "" {
				phash, _ := utils.HashPassword(tt.mockResponse.PasswordHash)
				tt.mockResponse.PasswordHash = string(phash)
			}
			mockStorage.On("GetUser", mock.Anything, tt.requestBody.Login).Return(tt.mockResponse, tt.mockError)

			body, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(body))
			handler := handlers.LoginHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

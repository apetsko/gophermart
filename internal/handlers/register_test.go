package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/mocks"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHandler(t *testing.T) {
	mockStorage := new(mocks.Storage)
	mockHasher := new(mocks.PasswordHasher) // Используем мока хэширования
	logger, _ := logging.NewLogger(zapcore.DebugLevel)

	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
		Secret:  "test_secret",
	}

	cases := []struct {
		name           string
		requestBody    models.User
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Successful registration",
			requestBody:    models.User{Login: "new_user1", Password: "securepassword"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "User already exists",
			requestBody:    models.User{Login: "existing_user", Password: "password"},
			mockError:      models.ErrUserExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Storage error",
			requestBody:    models.User{Login: "fail_user", Password: "password"},
			mockError:      errors.New("database failure"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil
			mockHasher.ExpectedCalls = nil

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tt.requestBody.Password), bcrypt.DefaultCost)
			require.NoError(t, err)

			mockHasher.On("HashPassword", tt.requestBody.Password).Return(hashedPassword, nil)

			mockStorage.On("AddUser", mock.Anything, mock.MatchedBy(func(ue *models.UserEntry) bool {
				return ue.Username == tt.requestBody.Login
			})).Return(1, tt.mockError)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(body))
			w := httptest.NewRecorder()
			handler := handlers.RegisterHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			mockStorage.AssertExpectations(t)
		})
	}
}

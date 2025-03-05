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
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestRegisterHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
		Secret:  "test_secret",
	}

	cases := []struct {
		name           string
		requestBody    interface{}
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			requestBody:    "invalid_json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Successful registration",
			requestBody: models.User{
				Login:    "new_user",
				Password: "securepassword",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User already exists",
			requestBody: models.User{
				Login:    "existing_user",
				Password: "password",
			},
			mockError:      models.ErrUserExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Storage error",
			requestBody: models.User{
				Login:    "fail_user",
				Password: "password",
			},
			mockError:      errors.New("database failure"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil

			if _, ok := tc.requestBody.(models.User); ok {
				mockStorage.On("AddUser", mock.Anything, mock.Anything).Return(1, tc.mockError)
			}

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			w := httptest.NewRecorder()
			handler := handlers.RegisterHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

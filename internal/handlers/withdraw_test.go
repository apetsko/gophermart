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

func TestWithdrawHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
	}

	cases := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Missing userID",
			userID:         "",
			requestBody:    models.Withdraw{Order: "123456789", Sum: 100},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid userID",
			userID:         "invalid_id",
			requestBody:    models.Withdraw{Order: "123456789", Sum: 100},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid JSON",
			userID:         "1",
			requestBody:    "invalid_json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "Successful withdrawal",
			userID: "1",
			requestBody: models.Withdraw{
				Order: "79927398713",
				Sum:   50,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Invalid Luhn order number",
			userID: "1",
			requestBody: models.Withdraw{
				Order: "123456789",
				Sum:   50,
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "Insufficient funds",
			userID: "1",
			requestBody: models.Withdraw{
				Order: "79927398713",
				Sum:   1000,
			},
			mockError:      models.ErrInsufficientFunds,
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:   "Storage error",
			userID: "1",
			requestBody: models.Withdraw{
				Order: "79927398713",
				Sum:   100,
			},
			mockError:      errors.New("storage error"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil

			if wd, ok := tc.requestBody.(models.Withdraw); ok {
				mockStorage.On("Withdraw", mock.Anything, mock.Anything, wd, mock.Anything).Return(&models.UserBalance{}, tc.mockError)
			}

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/withdraw", bytes.NewReader(body))
			if tc.userID != "" {
				req.Header.Set("userID", tc.userID)
			}
			w := httptest.NewRecorder()
			handler := handlers.WithdrawHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

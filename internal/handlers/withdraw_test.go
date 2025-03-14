package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/mocks"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestWithdrawHandler(t *testing.T) {
	mockStorage := new(mocks.Storage)
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
	}

	tests := []struct {
		name           string
		userID         string
		Request        models.WithdrawRequest
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Missing userID",
			userID:         "",
			Request:        models.WithdrawRequest{Order: "123456789", Sum: 100},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid userID",
			userID:         "invalid_id",
			Request:        models.WithdrawRequest{Order: "123456789", Sum: 100},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Successful withdrawal",
			userID: "1",
			Request: models.WithdrawRequest{
				Order: "79927398713",
				Sum:   50,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "Invalid Luhn order number",
			userID: "1",
			Request: models.WithdrawRequest{
				Order: "123456789",
				Sum:   50,
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "Insufficient funds",
			userID: "1",
			Request: models.WithdrawRequest{
				Order: "79927398713",
				Sum:   1000,
			},
			mockError:      models.ErrInsufficientFunds,
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:   "Storage error",
			userID: "1",
			Request: models.WithdrawRequest{
				Order: "79927398713",
				Sum:   100,
			},
			mockError:      errors.New("storage error"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil

			wd := models.Withdraw{
				Order:    tt.Request.Order,
				SumMinor: int64(math.Round(tt.Request.Sum * 100)),
			}

			mockStorage.On("Withdraw",
				mock.Anything,
				mock.Anything,
				wd,
				mock.Anything).
				Return(&models.UserBalance{}, tt.mockError)

			body, _ := json.Marshal(tt.Request)
			req := httptest.NewRequest(http.MethodPost, "/api/user/withdraw", bytes.NewReader(body))
			if tt.userID != "" {
				req.Header.Set("userID", tt.userID)
			}
			w := httptest.NewRecorder()
			handler := handlers.WithdrawHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

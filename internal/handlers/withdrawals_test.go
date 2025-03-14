package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/mocks"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestWithdrawalsHandler(t *testing.T) {
	mockStorage := new(mocks.Storage)
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
	}

	now := time.Now()
	tests := []struct {
		name           string
		userID         string
		mockResponse   []models.Withdraw
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Missing userID",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid userID",
			userID:         "invalid_id",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "No withdrawals found",
			userID:         "1",
			mockResponse:   nil,
			mockError:      models.ErrWithdrawalsNotFound,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "Successful withdrawals retrieval",
			userID: "1",
			mockResponse: []models.Withdraw{
				{Order: "79927398713", SumMinor: 5000, ProcessedAt: &now},
				{Order: "123456789", SumMinor: 2000, ProcessedAt: &now},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Storage error",
			userID:         "1",
			mockError:      errors.New("storage error"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil

			mockStorage.On("Withdrawals",
				mock.Anything,
				mock.Anything).
				Return(tt.mockResponse, tt.mockError)

			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			if tt.userID != "" {
				req.Header.Set("userID", tt.userID)
			}
			w := httptest.NewRecorder()
			handler := handlers.WithdrawalsHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				var withdrawals []models.Withdraw
				err := json.NewDecoder(resp.Body).Decode(&withdrawals)
				assert.NoError(t, err)

				assert.Equal(t, len(tt.mockResponse), len(withdrawals))

				for i := range tt.mockResponse {
					assert.Equal(t, tt.mockResponse[i].Order, withdrawals[i].Order)
					assert.Equal(t, tt.mockResponse[i].SumMinor/100, withdrawals[i].SumMinor)
					assert.WithinDuration(t, *tt.mockResponse[i].ProcessedAt, *withdrawals[i].ProcessedAt, time.Millisecond)
				}
			}
		})
	}
}

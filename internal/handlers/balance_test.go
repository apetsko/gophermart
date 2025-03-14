package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/mocks" // Новый мок
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestBalanceHandler(t *testing.T) {
	mockStorage := new(mocks.Storage) // Используем мок, созданный mockery
	logger, _ := logging.NewLogger(zapcore.FatalLevel)
	h := &handlers.URLHandler{
		Storage: mockStorage,
		Logger:  logger,
	}

	tests := []struct {
		name       string
		userID     string
		mockResp   *models.UserBalance
		mockErr    error
		expectCode int
	}{
		{"valid user, returns balance", "1", &models.UserBalance{CurrentMinor: 10005, WithdrawnMinor: 5010}, nil, http.StatusOK},
		{"missing userID header", "", nil, nil, http.StatusUnauthorized},
		{"invalid userID", "invalid", nil, nil, http.StatusUnauthorized},
		{"storage error", "2", nil, models.ErrBalanceNotFound, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil

			r := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			if tt.userID != "" {
				r.Header.Set("userID", tt.userID)
			}
			w := httptest.NewRecorder()

			if tt.userID != "" {
				userID, err := strconv.ParseInt(tt.userID, 10, 64)
				if err == nil {
					mockStorage.On("Balance", mock.Anything, userID).Return(tt.mockResp, tt.mockErr)
				}
			}
			var mockBalanceResp models.UserBalanceResponse
			if tt.mockResp != nil {
				mockBalanceResp = models.UserBalanceResponse{
					Current:   float64(tt.mockResp.CurrentMinor) / 100.0,
					Withdrawn: float64(tt.mockResp.WithdrawnMinor) / 100.0,
				}
			}
			handlers.BalanceHandler(h)(w, r)

			assert.Equal(t, tt.expectCode, w.Code)

			if tt.expectCode == http.StatusOK {
				var resp models.UserBalanceResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, mockBalanceResp, resp)
			}

			mockStorage.AssertExpectations(t) // Проверяем, что вызовы соответствуют ожиданиям
		})
	}
}

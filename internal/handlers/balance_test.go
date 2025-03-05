package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/apetsko/gophermart/internal/handlers"
	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestBalanceHandler(t *testing.T) {
	mockStorage := new(MockStorage)
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
		{"valid user, returns balance", "1", &models.UserBalance{Current: 100.5, Withdrawn: 50.0}, nil, http.StatusOK},
		{"missing userID header", "", nil, nil, http.StatusUnauthorized},
		{"invalid userID", "invalid", nil, nil, http.StatusUnauthorized},
		{"storage error", "2", nil, models.ErrBalanceNotFound, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil // Сброс прошлых вызовов

			r := httptest.NewRequest(http.MethodGet, "/balance", nil)
			if tt.userID != "" {
				r.Header.Set("userID", tt.userID)
			}
			w := httptest.NewRecorder()

			if tt.userID != "" {
				userID, err := strconv.ParseInt(tt.userID, 10, 64)
				if err == nil { // Только если userID корректный
					mockStorage.On("Balance", mock.Anything, userID).Return(tt.mockResp, tt.mockErr)
				}
			}

			handlers.BalanceHandler(h)(w, r)

			assert.Equal(t, tt.expectCode, w.Code)

			if tt.expectCode == http.StatusOK {
				var resp models.UserBalance
				err := json.NewDecoder(w.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, *tt.mockResp, resp)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

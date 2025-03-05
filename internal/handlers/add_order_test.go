package handlers_test

import (
	"bytes"
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

func TestAddOrderHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		order          string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Valid order",
			userID:         "1",
			order:          "1234567812345670", // Luhn-valid
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Invalid order (Luhn check fails)",
			userID:         "1",
			order:          "1234567812345678", // Luhn-invalid
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Order already exists",
			userID:         "1",
			order:          "1234567812345670",
			mockError:      models.ErrOrderAlreadyExists,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Order exists for another user",
			userID:         "1",
			order:          "1234567812345670",
			mockError:      models.ErrOrderExistsForAnotherUser,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Unauthorized (missing userID)",
			userID:         "",
			order:          "1234567812345670",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty request body",
			userID:         "1",
			order:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Storage error",
			userID:         "1",
			order:          "1234567812345670",
			mockError:      errors.New("some internal error"),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger, _ := logging.NewLogger(zapcore.DebugLevel)
			mockStorage := new(MockStorage)

			h := &handlers.URLHandler{
				Storage: mockStorage,
				Logger:  logger,
			}

			if tc.mockError != nil {
				mockStorage.On("AddOrder", mock.Anything, mock.AnythingOfType("int64"), tc.order).
					Return(tc.mockError)
			} else {
				mockStorage.On("AddOrder", mock.Anything, mock.AnythingOfType("int64"), tc.order).
					Return(nil)
			}

			req := httptest.NewRequest(http.MethodPost, "/add_order", bytes.NewBufferString(tc.order))
			if tc.userID != "" {
				req.Header.Set("userID", tc.userID)
			}
			w := httptest.NewRecorder()
			handlers.AddOrderHandler(h)(w, req)

			resp := w.Result()
			defer resp.Body.Close()
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

package handlers_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestListOrdersHandler(t *testing.T) {
	mockStorage := new(mocks.Storage)
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
	}

	fixedTime := time.Date(2025, 3, 5, 12, 0, 0, 0, time.UTC)
	accVal := int64(123)
	tests := []struct {
		name           string
		userIDHeader   string
		mockResponse   []models.UserOrderEntry
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing userID header",
			userIDHeader:   "",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "Invalid userID",
			userIDHeader:   "abc",
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "",
		},
		{
			name:           "No orders found",
			userIDHeader:   "1",
			mockResponse:   nil,
			mockError:      models.ErrOrderNotFound,
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
		{
			name:           "Internal server error",
			userIDHeader:   "1",
			mockResponse:   nil,
			mockError:      errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
		{
			name:         "Successful request",
			userIDHeader: "1",
			mockResponse: []models.UserOrderEntry{
				{
					Number:       "12345",
					Status:       "PROCESSED",
					AccrualMinor: &accVal,
					UploadedAt:   fixedTime,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"accrual":1.23,"number":"12345","status":"PROCESSED","uploaded_at":"2025-03-05T12:00:00Z"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil

			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if tt.userIDHeader != "" {
				req.Header.Set("userID", tt.userIDHeader)

				userID, err := strconv.ParseInt(tt.userIDHeader, 10, 64)
				if err == nil {
					mockStorage.On("ListOrders", mock.Anything, userID).Return(tt.mockResponse, tt.mockError)
				}
			}

			w := httptest.NewRecorder()
			handler := handlers.ListOrdersHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, string(body))
			} else {
				assert.Equal(t, tt.expectedBody, string(body))
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

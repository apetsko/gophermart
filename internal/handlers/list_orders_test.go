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
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

func TestListOrdersHandler(t *testing.T) {
	mockStorage := new(MockStorage)
	logger, _ := logging.NewLogger(zapcore.DebugLevel)
	h := &handlers.URLHandler{
		Logger:  logger,
		Storage: mockStorage,
	}

	fixedTime := time.Date(2025, 3, 5, 12, 0, 0, 0, time.UTC)

	cases := []struct {
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
					Number:     "12345",
					Status:     "PROCESSED",
					Accrual:    nil,
					UploadedAt: fixedTime,
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"number":"12345","status":"PROCESSED","uploaded_at":"2025-03-05T12:00:00Z"}]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil // Очистка предыдущих моков
			if tc.userIDHeader != "" {
				userID, _ := strconv.ParseInt(tc.userIDHeader, 10, 64)
				mockStorage.On("ListOrders", mock.Anything, userID).Return(tc.mockResponse, tc.mockError)
			}

			req := httptest.NewRequest(http.MethodGet, "/orders", nil)
			if tc.userIDHeader != "" {
				req.Header.Set("userID", tc.userIDHeader)
			}

			w := httptest.NewRecorder()
			handler := handlers.ListOrdersHandler(h)
			handler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(body))
			} else {
				assert.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}

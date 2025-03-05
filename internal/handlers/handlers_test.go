package handlers_test

import (
	"context"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) AddOrder(ctx context.Context, userID int64, order string) error {
	args := m.Called(ctx, userID, order)
	return args.Error(0)
}

func (m *MockStorage) Balance(ctx context.Context, id int64) (*models.UserBalance, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBalance), args.Error(1)
}

func (m *MockStorage) ListOrders(ctx context.Context, userID int64) ([]models.UserOrderEntry, error) {
	args := m.Called(ctx, userID)
	if orders, ok := args.Get(0).([]models.UserOrderEntry); ok {
		return orders, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorage) GetUser(ctx context.Context, login string) (*models.UserEntry, error) {
	args := m.Called(ctx, login)
	if user, ok := args.Get(0).(models.UserEntry); ok {
		return &user, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorage) AddUser(ctx context.Context, u *models.UserEntry) (int, error) {
	args := m.Called(ctx, u)
	return args.Int(0), args.Error(1)
}

func (m *MockStorage) Withdraw(ctx context.Context, id int64, wd models.Withdraw, logger logging.Logger) (*models.UserBalance, error) {
	args := m.Called(ctx, id, wd, logger)
	return args.Get(0).(*models.UserBalance), args.Error(1)
}

func (m *MockStorage) Withdrawals(ctx context.Context, id int64) ([]models.Withdraw, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]models.Withdraw), args.Error(1)
}

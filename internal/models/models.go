package models

import (
	"errors"
	"time"
)

var (
	ErrUserExists                = errors.New("user already exists")
	ErrUserNotFound              = errors.New("user not found")
	ErrOrderNotFound             = errors.New("order not found")
	ErrWithdrawalsNotFound       = errors.New("withdrawals not found")
	ErrOrderAlreadyExists        = errors.New("order already uploaded by this user")
	ErrOrderExistsForAnotherUser = errors.New("order already uploaded by another user")
	ErrInsufficientFunds         = errors.New("insufficient funds")
	ErrBalanceNotFound           = errors.New("balance not found")
)

// TODO REPLACE IT OR DELETE
type BatchDeleteRequest struct {
	Ids    []string
	UserID string
}

type User struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserEntry struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"password_hash"`
	Balance      float64 `json:"balance"`
}

type AccrualResponse struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

type UserOrderEntry struct {
	Number       string    `json:"number"`
	Status       string    `json:"status"`
	AccrualMinor *int64    `json:"accrual,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type UserOrderEntryResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type UserBalance struct {
	CurrentMinor   int64 `json:"current"`
	WithdrawnMinor int64 `json:"withdrawn"`
}

type UserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	Order       string     `json:"order"`
	SumMinor    int64      `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

type WithdrawRequest struct {
	Order       string     `json:"order"`
	Sum         float64    `json:"sum"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}

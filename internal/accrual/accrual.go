package accrual

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/storage/postgres"
	"github.com/go-resty/resty/v2"
)

type Order struct {
	ID          int64
	UserID      int64
	OrderNumber string
}

type OrderWorker struct {
	db     postgres.PgxPoolIface
	client *resty.Client
	apiURL string
	logger *logging.Logger
}

func NewOrderWorker(db postgres.PgxPoolIface, apiURL string, logger *logging.Logger) *OrderWorker {
	return &OrderWorker{
		db:     db,
		client: resty.New(),
		apiURL: apiURL,
		logger: logger,
	}
}

func ProcessAccrual(ctx context.Context, db postgres.PgxPoolIface, accrualURL string, numWorkers int, logger *logging.Logger) error {
	for i := 0; i < numWorkers; i++ {
		worker := NewOrderWorker(db, accrualURL, logger)
		go worker.Start(ctx)
	}
	return nil
}

func (w *OrderWorker) Start(ctx context.Context) {
	wait := time.Second
	ticker := time.NewTicker(wait)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("Order worker stopped")
			return
		case <-ticker.C:
			if err := w.processOrders(ctx); err != nil {
				w.logger.Error("Error processing orders:", err)
			}
		}
	}
}

func (w *OrderWorker) processOrders(ctx context.Context) error {
	orders, err := w.fetchOrders(ctx)
	if err != nil {
		return fmt.Errorf("fetch orders: %w", err)
	}

	for _, order := range orders {
		err := w.processOrder(ctx, order)
		if err != nil {
			w.logger.Errorf("Failed to process order %d: %v", order.ID, err)
			continue
		}
	}
	return nil
}

func (w *OrderWorker) fetchOrders(ctx context.Context) ([]Order, error) {
	query := `
		SELECT id, user_id, order_number
		FROM orders 
		WHERE status IN ('NEW', 'PROCESSING', 'REGISTERED')
		LIMIT 10`
	rows, err := w.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.OrderNumber); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (w *OrderWorker) processOrder(ctx context.Context, order Order) error {
	url := fmt.Sprintf("%s/api/orders/%s", w.apiURL, order.OrderNumber)

	resp, err := w.client.R().
		SetContext(ctx).
		SetResult(&models.AccrualResponse{}).
		Get(url)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return w.handleSuccessResponse(ctx, order, resp.Result().(*models.AccrualResponse))
	case http.StatusNoContent:
		return w.updateOrderStatus(ctx, order.ID, "INVALID")
	case http.StatusTooManyRequests:
		retryAfter := resp.Header().Get("Retry-After")
		w.logger.Debugf("Rate limited. Retry after %s seconds", retryAfter)
		return errors.New("rate limited")
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}
}

func (w *OrderWorker) handleSuccessResponse(ctx context.Context, order Order, accrualResp *models.AccrualResponse) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				w.logger.Errorf("Failed to rollback transaction: %v", err)
			}
		}
	}()

	status := accrualResp.Status
	var accrualValue float64

	if accrualResp.Accrual != nil {
		accrualValue = *accrualResp.Accrual
	}

	const updateOrders = `UPDATE orders SET status = $1, accrual = $2 WHERE id = $3`
	if _, err := tx.Exec(ctx, updateOrders, status, accrualValue, order.ID); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	const increaseBalance = `
    UPDATE users 
    SET current = current + $1
    WHERE id = $2
    RETURNING current;
`
	var current float64
	err = tx.QueryRow(ctx, increaseBalance, accrualValue, order.UserID).Scan(&current)
	if err != nil {
		return fmt.Errorf("failed to update balance. UserID: %d, %w", order.UserID, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	w.logger.Debugf("Order %d updated: status=%s, accrual=%.2f", order.UserID, status, accrualValue)
	return nil
}

func (w *OrderWorker) updateOrderStatus(ctx context.Context, orderID int64, status string) error {
	const updateOrders = `
		UPDATE orders
		SET status = $1
		WHERE id = $2;
`
	if _, err := w.db.Exec(ctx, updateOrders, status, orderID); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	w.logger.Debugf("Order %d updated: status=%s", orderID, status)
	return nil
}

package accrual

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/apetsko/gophermart/internal/storage/postgres"
	"resty.dev/v3"
)

type Order struct {
	ID          int64
	UserID      int64
	OrderNumber string
}

type OrderWorker struct {
	db     postgres.PgxPoolIface
	client *resty.Client
	logger *logging.Logger
}

func NewOrderWorker(db postgres.PgxPoolIface, client *resty.Client, logger *logging.Logger) *OrderWorker {
	return &OrderWorker{
		db:     db,
		client: client,
		logger: logger,
	}
}

func ProcessAccrual(ctx context.Context, db postgres.PgxPoolIface, accrualURL string, numWorkers int, logger *logging.Logger) {
	logger.Debugf("running accrual process with %d workers", numWorkers)
	client := resty.New()
	client.SetBaseURL(accrualURL)

	for i := 0; i < numWorkers; i++ {
		worker := NewOrderWorker(db, client, logger)
		go worker.Start(ctx, time.Second)
	}
}

func (w *OrderWorker) Start(ctx context.Context, wait time.Duration) {
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
		return fmt.Errorf("failed fetch orders: %w", err)
	}

	for _, order := range orders {
		err := w.processOrder(ctx, order)
		if err != nil {
			w.logger.Errorf("failed to process order %d: %v", order.ID, err)
			continue
		}
	}
	return nil
}

func (w *OrderWorker) fetchOrders(ctx context.Context) ([]Order, error) {
	query := `
	WITH selected_orders AS (
		SELECT id, user_id, order_number
		FROM orders
		WHERE status NOT IN ('INVALID', 'PROCESSED')
		AND start_process_at < NOW() - INTERVAL '5 minutes'
		ORDER BY start_process_at ASC
		LIMIT 10
		FOR UPDATE SKIP LOCKED
	)
	UPDATE orders
	SET start_process_at = NOW() + INTERVAL '3 minutes'
	FROM selected_orders
	WHERE orders.id = selected_orders.id
	RETURNING orders.id, orders.user_id, orders.order_number;`

	rows, err := w.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.OrderNumber); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}

func (w *OrderWorker) processOrder(ctx context.Context, order Order) error {
	url := fmt.Sprintf("/api/orders/%s", order.OrderNumber)
	var accResp models.AccrualResponse
	resp, err := w.client.R().
		SetContext(ctx).
		SetResult(&accResp).
		Get(url)

	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		return w.handleSuccessResponse(ctx, order, &accResp)
	case http.StatusNoContent:
		return w.updateOrderStatus(ctx, order.ID, "INVALID")
	case http.StatusTooManyRequests:
		retryAfter := resp.Header().Get("Retry-After")
		w.logger.Debugf("Order %s. rate limited. Retry after %s seconds", order.ID, retryAfter)
		return w.updateOrderTime(ctx, order.ID, retryAfter)
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
	var accrualMinor int

	if accrualResp.Accrual != nil {
		var kopecks float64 = 100
		accrualMinor = int(math.Round(*accrualResp.Accrual * kopecks))
	}

	const updateOrders = `UPDATE orders SET status = $1, accrual_minor = $2 WHERE id = $3`
	if _, err := tx.Exec(ctx, updateOrders, status, accrualMinor, order.ID); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	const lockBalance = `SELECT current_minor FROM users WHERE id = $1 FOR UPDATE;`
	if _, err = tx.Exec(ctx, lockBalance, order.UserID); err != nil {
		return fmt.Errorf("failed to lock balance. UserID: %d, %w", order.UserID, err)
	}

	const increaseBalance = `
    UPDATE users 
    SET current_minor = users.current_minor + $1
    WHERE id = $2
    RETURNING current_minor;
`

	if _, err := tx.Exec(ctx, increaseBalance, accrualMinor, order.UserID); err != nil {
		return fmt.Errorf("failed to update balance. UserID: %d, %w", order.UserID, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	w.logger.Debugf("Order %d updated: status=%s, accrual=%.2f", order.UserID, status, float64(accrualMinor)/100)
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

func (w *OrderWorker) updateOrderTime(ctx context.Context, orderID int64, retryAfter string) error {
	var startedAt time.Time
	now := time.Now()

	if retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			startedAt = now.Add(-5 * time.Minute).Add(time.Duration(seconds) * time.Second)
		} else if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
			startedAt = t
		} else {
			w.logger.Warnf("Invalid retryAfter format: %s", retryAfter)
			startedAt = now
		}
	} else {
		startedAt = now
	}

	const updateOrders = `
		UPDATE orders
		SET start_process_at = $1
		WHERE id = $2;
	`

	if _, err := w.db.Exec(ctx, updateOrders, startedAt, orderID); err != nil {
		return fmt.Errorf("failed to update order started_at: %w", err)
	}
	w.logger.Debugf("Order %d updated: started_at=%s", orderID, startedAt)
	return nil
}

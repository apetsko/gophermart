package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/apetsko/gophermart/internal/logging"
	"github.com/apetsko/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgxPoolIface interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Begin(context.Context) (pgx.Tx, error)
	Close()
}

//go:embed migrations/*.sql
var migrations embed.FS

type Storage struct {
	DB PgxPoolIface
}

func migrate(conn string, logger *logging.Logger) error {
	goose.SetLogger(logger)
	goose.SetBaseFS(migrations)
	db, err := sql.Open("pgx", conn)
	if err != nil {
		return fmt.Errorf("goose: failed to open DB: %w", err)
	}
	defer db.Close()

	err = goose.Up(db, "migrations")
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func New(conn string, logger *logging.Logger) (*Storage, error) {
	if err := migrate(conn, logger); err != nil {
		return nil, err
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return &Storage{DB: pool}, nil
}

func (p *Storage) Close() error {
	p.DB.Close()
	return nil
}

func (p *Storage) AddUser(ctx context.Context, u *models.UserEntry) (int, error) {
	const insert = `
		INSERT INTO users (username, password_hash, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (username) DO NOTHING
		RETURNING id;
	`

	var id int

	err := p.DB.QueryRow(ctx, insert, u.Username, u.PasswordHash).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, models.ErrUserExists // ✅ Возвращаем осмысленную ошибку
	}

	return id, err
}

func (p *Storage) GetUser(ctx context.Context, login string) (*models.UserEntry, error) {
	const getUserHash = `
		SELECT id, username, password_hash FROM users
		WHERE username = $1;
	`

	var u models.UserEntry

	err := p.DB.QueryRow(ctx, getUserHash, login).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (p *Storage) AddOrder(ctx context.Context, userID int64, order string) error {
	const check = `SELECT user_id FROM orders WHERE order_number = $1`
	var existingUserID int64
	err := p.DB.QueryRow(ctx, check, order).Scan(&existingUserID)

	if err == nil {
		if existingUserID == userID {
			return models.ErrOrderAlreadyExists
		}
		return models.ErrOrderExistsForAnotherUser
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	const insert = `
		INSERT INTO orders (user_id, order_number, status, uploaded_at, start_process_at)
		VALUES ($1, $2, 'NEW', NOW(), NOW() - INTERVAL '5 MINUTE')
	`
	_, err = p.DB.Exec(ctx, insert, userID, order)
	return err
}

func (p *Storage) ListOrders(ctx context.Context, userID int64) (ee []models.UserOrderEntry, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	const query = "SELECT order_number, status, accrual, uploaded_at FROM orders WHERE user_id = $1"

	rows, err := p.DB.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	found := false

	for rows.Next() {
		var entry = new(models.UserOrderEntry)
		if err := rows.Scan(&entry.Number, &entry.Status, &entry.Accrual, &entry.UploadedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		ee = append(ee, *entry)
		found = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	if !found {
		return nil, models.ErrOrderNotFound
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return ee, nil
}

func (p *Storage) Balance(ctx context.Context, id int64) (*models.UserBalance, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	const query = "SELECT current, withdrawn FROM users WHERE id = $1"

	b := new(models.UserBalance)
	if err := p.DB.QueryRow(ctx, query, id).Scan(&b.Current, &b.Withdrawn); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return b, nil
}

func (p *Storage) Withdraw(ctx context.Context, userID int64, wd models.Withdraw, logger logging.Logger) (*models.UserBalance, error) {
	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(ctx); err != nil {
				logger.Errorf("Failed to rollback transaction: %v", err)
			}
		}
	}()

	const getBalance = `SELECT current FROM users WHERE id = $1;`
	var balance float64
	err = tx.QueryRow(ctx, getBalance, userID).Scan(&balance)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance < wd.Sum {
		return nil, models.ErrInsufficientFunds
	}

	const addToWithdrawals = `
		INSERT INTO withdrawals (user_id, order_number, sum, processed_at)
		VALUES ($1, $2, $3, NOW());`
	_, err = tx.Exec(ctx, addToWithdrawals, userID, wd.Order, wd.Sum)
	if err != nil {
		return nil, fmt.Errorf("failed to insert withdrawal: %w", err)
	}

	const decreaseBalance = `
		UPDATE users 
		SET current = current - $2, withdrawn = withdrawn + $2 
		WHERE id = $1 
		RETURNING current, withdrawn;`
	b := new(models.UserBalance)
	err = tx.QueryRow(ctx, decreaseBalance, userID, wd.Sum).Scan(&b.Current, &b.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return b, nil
}

func (p *Storage) Withdrawals(ctx context.Context, userID int64) ([]models.Withdraw, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	const withdrawal = `
		SELECT order_number, sum, processed_at  
		FROM withdrawals WHERE user_id = $1
		`

	rows, err := p.DB.Query(ctx, withdrawal, userID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var ww []models.Withdraw

	for rows.Next() {
		var w models.Withdraw
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		ww = append(ww, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	if len(ww) == 0 {
		return nil, models.ErrWithdrawalsNotFound
	}
	return ww, nil
}

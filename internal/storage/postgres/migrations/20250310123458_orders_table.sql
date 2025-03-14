-- +goose Up
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    order_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'NEW',
    accrual_minor BIGINT DEFAULT 0,
    uploaded_at TIMESTAMP DEFAULT NOW(),
    start_process_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_orders_status_process_time ON orders (status, start_process_at);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_status_process_time;
DROP TABLE IF EXISTS orders;



CREATE TABLE withdraws (
    id SERIAL PRIMARY KEY,
    withdraw  NUMERIC NOT NULL,
    order_num VARCHAR(255) UNIQUE NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX idx_order_num on withdraws(order_num);
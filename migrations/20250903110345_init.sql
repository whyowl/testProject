-- +goose Up
CREATE TABLE wallets (
                       wallet_id UUID PRIMARY KEY,
                       balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS wallets;

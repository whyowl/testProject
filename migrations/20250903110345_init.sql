-- +goose Up
CREATE TABLE wallets (
                       wallet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       balance INT NOT NULL DEFAULT 0,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS wallets;

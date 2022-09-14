-- +goose Up
-- +goose StatementBegin
CREATE TABLE wallets (
    id BIGSERIAL NOT NULL CONSTRAINT wallet_pk PRIMARY KEY,
    player_name VARCHAR UNIQUE NOT NULL,
    currency VARCHAR NOT NULL,
    balance BIGINT NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE wallets;
-- +goose StatementEnd

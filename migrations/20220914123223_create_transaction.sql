-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
    id VARCHAR NOT NULL CONSTRAINT transactions_pk PRIMARY KEY,
    player_name VARCHAR NOT NULL,
    withdraw BIGINT,
    deposit BIGINT,
    currency VARCHAR NOT NULL,
    balance_after_commit BIGINT,
    external_id VARCHAR NOT NULL UNIQUE,
    rolled_back BOOLEAN NOT NULL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions;
-- +goose StatementEnd

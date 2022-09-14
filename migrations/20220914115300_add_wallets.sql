-- +goose Up
-- +goose StatementBegin
INSERT INTO wallets (player_name, currency, balance)
VALUES
    ('user1', 'USD', 1000),
    ('user2', 'USD', 200),
    ('user3', 'EUR', 999),
    ('user4', 'RUB', 12000);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM wallets WHERE player_name IN ('user1', 'user2', 'user3', 'user4');
-- +goose StatementEnd

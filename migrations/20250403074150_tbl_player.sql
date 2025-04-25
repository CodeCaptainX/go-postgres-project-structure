-- +goose Up
CREATE TABLE tbl_players (
    id SERIAL PRIMARY KEY,
    suit_name VARCHAR(50) NOT NULL,
    suit_symbol VARCHAR(10) NOT NULL,
    status_id INT NOT NULL DEFAULT 1,
    "order" INT NOT NULL DEFAULT 1,
    created_by INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by INT,
    updated_at TIMESTAMP,
    deleted_by INT,
    deleted_at TIMESTAMP
);
-- +goose StatementBegin
INSERT INTO tbl_players (
    suit_name, 
    suit_symbol, 
    status_id, 
    "order", 
    created_by, 
    created_at
    )
VALUES
    ('heart', '♥️', 1, 1, 1, CURRENT_TIMESTAMP),
    ('diamond', '♦️', 1, 1, 1, CURRENT_TIMESTAMP),
    ('club', '♣️', 1, 1, 1, CURRENT_TIMESTAMP),
    ('spade', '♠️', 1, 1, 1, CURRENT_TIMESTAMP),
    ('joker', 'joker', 1, 1, 1, CURRENT_TIMESTAMP);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS tbl_players;

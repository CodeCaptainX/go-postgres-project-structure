-- +goose Up
-- Enable uuid extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tbl_players (
    id SERIAL PRIMARY KEY,
    player_uuid UUID NOT NULL DEFAULT uuid_generate_v4(), -- Added UUID
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    user_name VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    login_session TEXT NULL,
    profile_photo TEXT NULL,
    user_alias VARCHAR(255) NULL,
    phone_number VARCHAR NULL,
    user_avatar_id INTEGER NULL,
    commission DECIMAL (10,2) NULL DEFAULT 0,
    last_access TIMESTAMP WITHOUT TIME ZONE,
    status_id SMALLINT DEFAULT 1,
    "order" INTEGER NULL DEFAULT 1,
    created_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by INTEGER,
    updated_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by INTEGER,
    deleted_at TIMESTAMP WITHOUT TIME ZONE
);

-- +goose StatementBegin
INSERT INTO tbl_players (
    player_uuid, first_name, last_name, user_name, password, email, login_session, 
    profile_photo, user_alias, phone_number, user_avatar_id, 
    commission, last_access, status_id, "order", 
    created_by, created_at, updated_by, updated_at, deleted_by, deleted_at
) VALUES
(
    uuid_generate_v4(), 'John', 'Doe', 'admin', '123', 'johndoe@example.com', NULL, 
    NULL, 'JD', '1234567890', NULL, 
    5.50, NOW(), 1, 1, 
    1, NOW(), NULL, NULL, NULL, NULL
),
(
    uuid_generate_v4(), 'Jane', 'Smith', 'janesmith', 'hashedpassword2', 'janesmith@example.com', NULL, 
    NULL, 'JS', '0987654321', NULL, 
    10.00, NOW(), 1, 2, 
    1, NOW(), NULL, NULL, NULL, NULL
),
(
    uuid_generate_v4(), 'Alice', 'Brown', 'alicebrown', 'hashedpassword3', 'alicebrown@example.com', NULL, 
    NULL, 'AB', '5551234567', NULL, 
    7.75, NOW(), 1, 3, 
    1, NOW(), NULL, NULL, NULL, NULL
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS tbl_players;

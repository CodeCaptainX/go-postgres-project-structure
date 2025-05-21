-- +goose Up
CREATE TABLE tbl_roles (
    id SERIAL PRIMARY KEY,
    user_role_uuid UUID NOT NULL UNIQUE,
    user_role_name VARCHAR NOT NULL,
    user_role_desc TEXT NOT NULL,
    status BOOLEAN NOT NULL,
    "order" INTEGER DEFAULT 1,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by INTEGER,
    updated_at TIMESTAMP,
    deleted_by INTEGER,
    deleted_at TIMESTAMP
);

-- +goose StatementBegin
INSERT INTO tbl_roles (
    user_role_uuid, user_role_name, user_role_desc, status, "order", created_by, created_at
) VALUES
    ('9a6f17b3-f2d1-4df4-8ade-d1c8fbebdb97', 'admin', 'Role Admin', true, 1, 1, NOW()),
    ('01918ff3-57fd-7c09-93a3-7077087550b0', 'moderator', 'Role Moderator', true, 1, 1, NOW()),
    ('01918fdb-0d16-7b77-b0c6-dc681aada863', 'operator', 'Role Operator', true, 1, 1, NOW());
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS tbl_roles;
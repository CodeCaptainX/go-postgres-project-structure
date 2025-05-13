-- +goose Up
-- USERS AUDIT TABLE
CREATE TABLE tbl_users_audits (
    id SERIAL PRIMARY KEY,
    user_audit_uuid UUID NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    user_audit_context TEXT NOT NULL,
    user_audit_desc TEXT NOT NULL,
    audit_type_id INTEGER NOT NULL,
    user_agent TEXT NOT NULL,
    operator TEXT NOT NULL,
    ip VARCHAR NOT NULL,
    status_id INTEGER NOT NULL,
    "order" INTEGER DEFAULT 1,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_by INTEGER,
    updated_at TIMESTAMP,
    deleted_by INTEGER,
    deleted_at TIMESTAMP
);
-- +goose StatementBegin
INSERT INTO tbl_users_audits  (
    user_audit_uuid, user_id, user_audit_context, user_audit_desc,
    audit_type_id, user_agent, operator, ip, status_id,
    "order", created_by, created_at, updated_by, updated_at, deleted_by, deleted_at
) VALUES 
(
    'a1a1a1a1-a1a1-4a1a-a1a1-a1a1a1a1a1a1', 1,
    'create_user', 'Created admin user ADMIN', 1,
    'PostgreSQL Client', 'system', '127.0.0.1', 1,
    1, 1, NOW(), NULL, NULL, NULL, NULL
),
(
    'b2b2b2b2-b2b2-4b2b-b2b2-b2b2b2b2b2b2', 2,
    'create_user', 'Created developer user IT', 1,
    'PostgreSQL Client', 'system', '127.0.0.1', 1,
    1, 1, NOW(), NULL, NULL, NULL, NULL
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS tbl_users_audits;
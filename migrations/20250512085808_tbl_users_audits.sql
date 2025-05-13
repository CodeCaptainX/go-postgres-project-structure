-- +goose Up
CREATE SEQUENCE IF NOT EXISTS tbl_users_audits_id_seq START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

CREATE TABLE IF NOT EXISTS tbl_users_audits (
    id BIGINT PRIMARY KEY DEFAULT nextval('tbl_users_audits_id_seq'),
    player_uuid UUID NOT NULL DEFAULT uuid_generate_v4(), -- Added UUID column
    user_id BIGINT NOT NULL,
    audit_context TEXT NOT NULL,
    audit_desc TEXT NOT NULL,
    audit_type_id INT NOT NULL,
    user_agent TEXT,
    operator TEXT,
    ip VARCHAR(64),
    status_id SMALLINT DEFAULT 1,
    "order" BIGINT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_by BIGINT,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_by BIGINT,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- +goose Down
DROP TABLE IF EXISTS tbl_users_audits;
DROP SEQUENCE IF EXISTS tbl_users_audits_id_seq;
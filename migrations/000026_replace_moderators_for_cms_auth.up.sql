DROP TABLE IF EXISTS moderators;

CREATE TABLE IF NOT EXISTS moderator_roles (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(40) NOT NULL UNIQUE,
    name VARCHAR(80) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT moderator_roles_code_format CHECK (code ~ '^[a-z][a-z0-9_]*$')
);

INSERT INTO moderator_roles (code, name, description)
VALUES
    ('super_admin', 'Super Admin', 'Full CMS access, including future moderator and role management.'),
    ('admin', 'Admin', 'Standard CMS administration access.')
ON CONFLICT (code) DO NOTHING;

CREATE TABLE IF NOT EXISTS moderators (
    id BIGSERIAL PRIMARY KEY,
    login_id CHAR(8) NOT NULL UNIQUE,
    role_id BIGINT NOT NULL REFERENCES moderator_roles(id),
    display_name VARCHAR(120) NOT NULL,
    email VARCHAR,
    totp_secret_encrypted TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT moderators_login_id_format CHECK (login_id ~ '^[A-Za-z0-9]{8}$')
);

CREATE INDEX IF NOT EXISTS idx_moderators_role_id ON moderators(role_id);
CREATE INDEX IF NOT EXISTS idx_moderators_active_login_id ON moderators(login_id) WHERE is_active = TRUE AND deleted_at IS NULL;

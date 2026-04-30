CREATE TABLE IF NOT EXISTS user_showroom_relations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    showroom_id BIGINT NOT NULL REFERENCES showrooms(id),
    role_id BIGINT NOT NULL REFERENCES user_roles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE (user_id, showroom_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_showroom_relations_user_id ON user_showroom_relations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_showroom_relations_showroom_id ON user_showroom_relations(showroom_id);
CREATE INDEX IF NOT EXISTS idx_user_showroom_relations_role_id ON user_showroom_relations(role_id);

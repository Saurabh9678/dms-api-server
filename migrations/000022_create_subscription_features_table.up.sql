CREATE TYPE feature_value_type AS ENUM (
    'BOOLEAN',
    'NUMBER',
    'STRING'
);

CREATE TABLE IF NOT EXISTS subscription_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    value_type feature_value_type NOT NULL,
    category VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

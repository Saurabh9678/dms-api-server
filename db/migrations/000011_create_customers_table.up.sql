CREATE TABLE IF NOT EXISTS customers (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR NOT NULL,
    last_name VARCHAR,
    email VARCHAR,
    phone_number VARCHAR NOT NULL UNIQUE,
    alt_phone_number VARCHAR,
    address TEXT,
    city VARCHAR,
    state VARCHAR,
    pincode VARCHAR,
    id_proof_type VARCHAR,
    id_proof_number VARCHAR,
    id_proof_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

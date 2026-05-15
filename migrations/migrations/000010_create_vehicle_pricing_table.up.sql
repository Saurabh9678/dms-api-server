CREATE TABLE IF NOT EXISTS vehicle_pricing (
    id BIGSERIAL PRIMARY KEY,
    vehicle_id BIGINT UNIQUE NOT NULL REFERENCES vehicles(id),
    buying_price NUMERIC(10,2) NOT NULL,
    buying_date DATE NOT NULL,
    price_tag NUMERIC(10,2),
    tagged_at TIMESTAMPTZ,
    currency VARCHAR(10) NOT NULL DEFAULT 'inr',
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

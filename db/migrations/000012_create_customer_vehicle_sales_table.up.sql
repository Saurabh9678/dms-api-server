CREATE TABLE IF NOT EXISTS customer_vehicle_sales (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    vehicle_id BIGINT NOT NULL REFERENCES vehicles(id),
    sale_price NUMERIC(10,2) NOT NULL,
    sale_date DATE NOT NULL,
    payment_mode VARCHAR,
    receipt_url TEXT,
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_customer_vehicle_sales_customer_id ON customer_vehicle_sales(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_vehicle_sales_vehicle_id ON customer_vehicle_sales(vehicle_id);

ALTER TABLE customer_vehicle_sales
    ADD COLUMN IF NOT EXISTS sold_by BIGINT REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS sold_by_name VARCHAR,
    ADD COLUMN IF NOT EXISTS sold_by_country_code VARCHAR,
    ADD COLUMN IF NOT EXISTS sold_by_phone_number VARCHAR;

CREATE INDEX IF NOT EXISTS idx_customer_vehicle_sales_sold_by
    ON customer_vehicle_sales(sold_by);

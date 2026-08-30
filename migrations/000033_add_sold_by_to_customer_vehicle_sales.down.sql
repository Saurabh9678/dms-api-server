DROP INDEX IF EXISTS idx_customer_vehicle_sales_sold_by;

ALTER TABLE customer_vehicle_sales
    DROP COLUMN IF EXISTS sold_by_phone_number,
    DROP COLUMN IF EXISTS sold_by_country_code,
    DROP COLUMN IF EXISTS sold_by_name,
    DROP COLUMN IF EXISTS sold_by;

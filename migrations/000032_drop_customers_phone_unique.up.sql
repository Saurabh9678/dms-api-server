-- Allow duplicate customer phone numbers: each sale stores its own customer record.
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_phone_number_key;

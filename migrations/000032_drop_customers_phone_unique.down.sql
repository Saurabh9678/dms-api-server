-- Restore unique phone constraint (may fail if duplicate phones already exist).
ALTER TABLE customers ADD CONSTRAINT customers_phone_number_key UNIQUE (phone_number);

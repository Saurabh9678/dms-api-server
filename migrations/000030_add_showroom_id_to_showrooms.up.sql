ALTER TABLE showrooms
    ADD COLUMN showroom_id VARCHAR(8);

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM showrooms WHERE id >= 2821109907456) THEN
        RAISE EXCEPTION 'showroom id is too large to backfill into 8 base36 characters';
    END IF;
END $$;

CREATE OR REPLACE FUNCTION dms_temp_to_base36(input_value BIGINT)
RETURNS TEXT
LANGUAGE plpgsql
AS $$
DECLARE
    chars CONSTANT TEXT := '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    value BIGINT := input_value;
    result TEXT := '';
BEGIN
    IF value = 0 THEN
        RETURN '0';
    END IF;

    WHILE value > 0 LOOP
        result := substr(chars, ((value % 36)::INT + 1), 1) || result;
        value := value / 36;
    END LOOP;

    RETURN result;
END;
$$;

UPDATE showrooms
SET showroom_id = LPAD(dms_temp_to_base36(id), 8, '0')
WHERE showroom_id IS NULL;

DROP FUNCTION dms_temp_to_base36(BIGINT);

ALTER TABLE showrooms
    ADD CONSTRAINT chk_showrooms_showroom_id_format
        CHECK (showroom_id ~ '^[A-Z0-9]{8}$'),
    ALTER COLUMN showroom_id SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_showrooms_showroom_id
    ON showrooms(showroom_id);

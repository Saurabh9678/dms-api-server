DROP INDEX IF EXISTS idx_showrooms_showroom_id;

ALTER TABLE showrooms
    DROP CONSTRAINT IF EXISTS chk_showrooms_showroom_id_format,
    DROP COLUMN IF EXISTS showroom_id;

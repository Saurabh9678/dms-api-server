INSERT INTO moderator_roles (code, name, description)
VALUES
    ('super_admin', 'Super Admin', 'Full CMS access, including future moderator and role management.'),
    ('admin', 'Admin', 'Standard CMS administration access.')
ON CONFLICT (code) DO NOTHING;

ALTER TABLE moderators
    ALTER COLUMN totp_secret_encrypted DROP NOT NULL;

ALTER TABLE moderators
    ADD COLUMN IF NOT EXISTS totp_enabled_at TIMESTAMPTZ;

INSERT INTO moderators (login_id, role_id, display_name)
SELECT
    SUBSTRING(UPPER(MD5(RANDOM()::TEXT || CLOCK_TIMESTAMP()::TEXT)) FROM 1 FOR 8),
    moderator_roles.id,
    'CMS Super Admin'
FROM moderator_roles
WHERE moderator_roles.code = 'super_admin'
  AND NOT EXISTS (
      SELECT 1
      FROM moderators
      JOIN moderator_roles existing_roles ON existing_roles.id = moderators.role_id
      WHERE existing_roles.code = 'super_admin'
        AND moderators.deleted_at IS NULL
  );

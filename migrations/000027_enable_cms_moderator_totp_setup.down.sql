DELETE FROM moderators
USING moderator_roles
WHERE moderators.role_id = moderator_roles.id
  AND moderator_roles.code = 'super_admin'
  AND moderators.display_name = 'CMS Super Admin'
  AND moderators.totp_secret_encrypted IS NULL
  AND moderators.totp_enabled_at IS NULL
  AND moderators.last_login_at IS NULL;

DELETE FROM moderators
WHERE totp_secret_encrypted IS NULL;

ALTER TABLE moderators
    ALTER COLUMN totp_secret_encrypted SET NOT NULL;

ALTER TABLE moderators
    DROP COLUMN IF EXISTS totp_enabled_at;

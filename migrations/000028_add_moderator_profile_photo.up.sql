ALTER TABLE moderators
    ADD COLUMN IF NOT EXISTS profile_photo_url TEXT;

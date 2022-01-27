ALTER TABLE highlights.guilds
    ADD COLUMN blocked_channels text[] NOT NULL DEFAULT '{}';

ALTER TABLE highlights.highlights
    RENAME channel_blocks TO blocked_channels;

ALTER TABLE highlights.highlights
    RENAME user_blocks TO blocked_users;
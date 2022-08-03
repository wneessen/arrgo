CREATE TABLE IF NOT EXISTS guild_prefs
(
    guild_id bigint                      NOT NULL REFERENCES guilds ON DELETE CASCADE,
    pref_key varchar(128)                NOT NULL,
    pref_val bytea                       NOT NULL,
    is_enc   bool                        NOT NULL DEFAULT false,
    version  INTEGER                     NOT NULL DEFAULT 1,
    ctime    timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    mtime    timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (guild_id, pref_key)
);
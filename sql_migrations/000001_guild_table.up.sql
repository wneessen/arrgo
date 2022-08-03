CREATE TABLE IF NOT EXISTS guilds
(
    id             bigserial PRIMARY KEY,
    guild_id       VARCHAR(32)                 NOT NULL UNIQUE,
    guild_name     VARCHAR(512)                NOT NULL,
    owner_id       VARCHAR(32)                 NOT NULL,
    joined_at      timestamp(0) with time zone NOT NULL,
    system_channel VARCHAR(32)                 NOT NULL,
    version        INTEGER                     NOT NULL DEFAULT 1,
    ctime          timestamp(0) with time zone NOT NULL DEFAULT now(),
    mtime          timestamp(0) with time zone NOT NULL DEFAULT now()
);
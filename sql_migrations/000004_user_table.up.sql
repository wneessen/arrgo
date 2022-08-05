CREATE TABLE IF NOT EXISTS users
(
    id      bigserial PRIMARY KEY,
    user_id VARCHAR(32)                 NOT NULL UNIQUE,
    enc_key BYTEA,
    version INTEGER                     NOT NULL DEFAULT 1,
    ctime   timestamp(0) with time zone NOT NULL DEFAULT now(),
    mtime   timestamp(0) with time zone NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_prefs
(
    user_id  bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    pref_key varchar(128)                NOT NULL,
    pref_val bytea                       NOT NULL,
    is_enc   bool                        NOT NULL DEFAULT false,
    version  INTEGER                     NOT NULL DEFAULT 1,
    ctime    timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    mtime    timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, pref_key)
);

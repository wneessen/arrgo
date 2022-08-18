CREATE TABLE IF NOT EXISTS user_ledger
(
    id        bigserial PRIMARY KEY,
    user_id   bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    emissary  varchar(32)                 NOT NULL,
    band      int                         NULL,
    rank      bigint                      NOT NULL,
    score     bigint                      NULL,
    next_rank bigint                      NULL,
    ctime     timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

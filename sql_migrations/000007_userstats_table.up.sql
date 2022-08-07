CREATE TABLE IF NOT EXISTS user_stats
(
    id            bigserial PRIMARY KEY,
    user_id       bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    gold          bigint                      NULL,
    doubloons     bigint                      NULL,
    ancient_coins bigint                      NULL,
    kraken        bigint                      NULL,
    megalodon     bigint                      NULL,
    chests        bigint                      NULL,
    ships         bigint                      NULL,
    vomit         bigint                      NULL,
    distance      bigint                      NULL,
    ctime         timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

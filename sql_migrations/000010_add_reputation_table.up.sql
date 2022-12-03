CREATE TABLE IF NOT EXISTS user_reputation
(
    id              bigserial PRIMARY KEY,
    user_id         bigint                      NOT NULL REFERENCES users ON DELETE CASCADE,
    emissary        varchar(32)                 NOT NULL,
    motto           varchar(255)                NOT NULL,
    rank            varchar(255)                NOT NULL,
    lvl             int                         NOT NULL,
    xp              int                         NOT NULL,
    next_lvl        int                         NOT NULL,
    xp_next_lvl     int                         NOT NULL,
    titlestotal     int                         NOT NULL,
    titlesunlocked  int                         NOT NULL,
    emblemstotal    int                         NOT NULL,
    emblemsunlocked int                         NOT NULL,
    itemstotal      int                         NOT NULL,
    itemsunlocked   int                         NOT NULL,
    ctime           timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS deeds
(
    id            bigserial PRIMARY KEY,
    deed_type     varchar(128)                NOT NULL,
    description   text                        NOT NULL,
    valid_from    timestamp(0) with time zone NOT NULL,
    valid_thru    timestamp(0) with time zone NOT NULL,
    reward_type   varchar(32)                 NOT NULL,
    reward_amount int                         NOT NULL,
    reward_icon   char(1)                     NOT NULL,
    image_url     varchar(1024)               NOT NULL,
    ctime         timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    UNIQUE (deed_type, valid_from, valid_thru)
);

CREATE TABLE IF NOT EXISTS trade_routes
(
    id           bigserial PRIMARY KEY,
    outpost      varchar(128)                NOT NULL UNIQUE,
    sought_after varchar(32)                 NOT NULL,
    surplus      varchar(32)                 NOT NULL,
    validthru    timestamp(0) with time zone NOT NULL,
    version      INTEGER                     NOT NULL DEFAULT 1,
    ctime        timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    mtime        timestamp(0) with time zone NOT NULL DEFAULT NOW()
);

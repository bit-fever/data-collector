-- ======================================================================
-- ===   Sql Script for Database : Data Store
-- ===
-- === Build : 317
-- ======================================================================

CREATE TABLE system_data_1m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    system_code    text               not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,system_code)
  );

-- ======================================================================

CREATE TABLE system_data_5m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    system_code    text               not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,system_code)
  );

-- ======================================================================

CREATE TABLE system_data_15m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    system_code    text               not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,system_code)
  );

-- ======================================================================

CREATE TABLE system_data_60m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    system_code    text               not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,system_code)
  );

-- ======================================================================

CREATE TABLE system_data_1440m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    system_code    text               not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,system_code)
  );

-- ======================================================================

CREATE TABLE user_data_1m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    product_id     int                not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,product_id)
  );

-- ======================================================================

CREATE TABLE user_data_5m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    product_id     int                not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,product_id)
  );

-- ======================================================================

CREATE TABLE user_data_15m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    product_id     int                not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,product_id)
  );

-- ======================================================================

CREATE TABLE user_data_60m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    product_id     int                not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,product_id)
  );

-- ======================================================================

CREATE TABLE user_data_1440m
  (
    time           timestamptz        not null,
    symbol         text               not null,
    product_id     int                not null,
    open           double precision   not null,
    high           double precision   not null,
    low            double precision   not null,
    close          double precision   not null,
    up_volume      int                not null,
    down_volume    int                not null,
    up_ticks       int                not null,
    down_ticks     int                not null,
    open_interest  int                not null,

    unique(time,symbol,product_id)
  );

-- ======================================================================

SELECT create_hypertable('system_data_1m',  by_range('time'));
SELECT create_hypertable('system_data_5m',  by_range('time'));
SELECT create_hypertable('system_data_15m', by_range('time'));
SELECT create_hypertable('system_data_60m', by_range('time'));

CREATE INDEX system_time_index_1m  ON system_data_1m (symbol, time DESC);
CREATE INDEX system_time_index_5m  ON system_data_5m (symbol, time DESC);
CREATE INDEX system_time_index_15m ON system_data_15m(symbol, time DESC);
CREATE INDEX system_time_index_60m ON system_data_60m(symbol, time DESC);


SELECT create_hypertable('user_data_1m',  by_range('time'));
SELECT create_hypertable('user_data_5m',  by_range('time'));
SELECT create_hypertable('user_data_15m', by_range('time'));
SELECT create_hypertable('user_data_60m', by_range('time'));

CREATE INDEX user_time_index_1m  ON user_data_1m (symbol, time DESC);
CREATE INDEX user_time_index_5m  ON user_data_5m (symbol, time DESC);
CREATE INDEX user_time_index_15m ON user_data_15m(symbol, time DESC);
CREATE INDEX user_time_index_60m ON user_data_60m(symbol, time DESC);
-- ======================================================================


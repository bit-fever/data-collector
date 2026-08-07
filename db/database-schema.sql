-- ======================================================================
-- ===   Sql Script for Database : Data Collector
-- ===
-- === Build : 317
-- ======================================================================

CREATE TABLE data_product
  (
    id                      int,
    username                varchar(32)   not null,
    connection_code         varchar(8)    not null,
    system_code             varchar(8)    not null,
    symbol                  varchar(16)   not null,
    supports_multiple_data  tinyint       not null,
    connected               tinyint       not null,
    timezone                varchar(32)   not null,
    status                  tinyint       not null,
    months                  varchar(16),
    rollover_trigger        varchar(16),
    trading_session_id      int           not null,
    trading_session_config  text          not null,

    primary key(id),
    unique(username,connection_code,symbol)
  )
 ENGINE = InnoDB ;

-- ======================================================================

CREATE TABLE data_block
  (
    id           int           auto_increment,
    system_code  varchar(8)    not null,
    root         varchar(16)   not null,
    symbol       varchar(16)   not null,
    status       tinyint       not null,
    global       tinyint       not null,
    data_from    int,
    data_to      int,
    progress     tinyint       not null,

    primary key(id)
  )
 ENGINE = InnoDB ;

-- ======================================================================

CREATE TABLE broker_product
  (
    id                  int,
    username            varchar(32)   not null,
    connection_code     varchar(8)    not null,
    symbol              varchar(16)   not null,
    name                varchar(64)   not null,
    point_value         double        not null,
    cost_per_operation  double        not null,
    currency_code       varchar(16)   not null,

    primary key(id),
    unique(username,connection_code,symbol)
  )
 ENGINE = InnoDB ;

CREATE INDEX broker_productIDX1 ON broker_product(username);

-- ======================================================================

CREATE TABLE outbox
  (
    id         int           auto_increment,
    timestamp  datetime      not null,
    exchange   varchar(64)   not null,
    uuid       varchar(64)   not null,
    payload    mediumblob    not null,
    size       int           not null,

    primary key(id)
  )
 ENGINE = InnoDB ;

-- ======================================================================

CREATE TABLE data_instrument
  (
    id                  int           auto_increment,
    data_block_id       int,
    data_product_id     int           not null,
    symbol              varchar(16)   not null,
    name                varchar(64)   not null,
    expiration_date     datetime,
    continuous          tinyint       not null,
    virtual_instrument  tinyint       not null,
    month               char(1),
    rollover_date       datetime,
    rollover_delta      double,
    rollover_status     tinyint       not null,

    primary key(id),
    unique(data_product_id,symbol),

    foreign key(data_block_id) references data_block(id),
    foreign key(data_product_id) references data_product(id)
  )
 ENGINE = InnoDB ;

-- ======================================================================

CREATE TABLE ingestion_job
  (
    id                  int           auto_increment,
    data_instrument_id  int           not null,
    data_block_id       int           not null,
    filename            varchar(64)   not null,
    records             int           not null,
    bytes               bigint        not null,
    timezone            varchar(64)   not null,
    parser              varchar(64)   not null,
    error               text,

    primary key(id),

    foreign key(data_instrument_id) references data_instrument(id),
    foreign key(data_block_id) references data_block(id)
  )
 ENGINE = InnoDB ;

-- ======================================================================

CREATE TABLE download_job
  (
    id                  int           auto_increment,
    data_instrument_id  int           not null,
    data_block_id       int           not null,
    status              tinyint       not null,
    load_from           int           not null,
    load_to             int           not null,
    priority            tinyint       not null,
    user_connection     varchar(64),
    product_timezone    varchar(64)   not null,
    session             text          not null,
    curr_day            int           not null,
    tot_days            int           not null,
    error               text,

    primary key(id),

    foreign key(data_instrument_id) references data_instrument(id),
    foreign key(data_block_id) references data_block(id)
  )
 ENGINE = InnoDB ;

-- ======================================================================

CREATE TABLE bias_analysis
  (
    id                  int           auto_increment,
    username            varchar(32)   not null,
    data_instrument_id  int           not null,
    broker_product_id   int           not null,
    name                varchar(64),
    notes               text,
    created_at          datetime      not null,
    updated_at          datetime,

    primary key(id),

    foreign key(data_instrument_id) references data_instrument(id),
    foreign key(broker_product_id) references broker_product(id)
  )
 ENGINE = InnoDB ;

CREATE INDEX bias_analysisIDX1 ON bias_analysis(username);

-- ======================================================================

CREATE TABLE bias_config
  (
    id                int            auto_increment,
    bias_analysis_id  int            not null,
    start_day         smallint       not null,
    start_slot        smallint       not null,
    end_day           smallint       not null,
    end_slot          smallint       not null,
    months            smallint       not null,
    excludes          varchar(255),
    operation         tinyint        not null,
    gross_profit      double,
    net_profit        double,

    primary key(id),

    foreign key(bias_analysis_id) references bias_analysis(id)
  )
 ENGINE = InnoDB ;

CREATE INDEX bias_configIDX1 ON bias_config(bias_analysis_id);

-- ======================================================================


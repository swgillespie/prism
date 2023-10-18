CREATE TABLE IF NOT EXISTS meta.table_schemas (
    table_name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    type INT4 NOT NULL,

    PRIMARY KEY (table_name, column_name)
);

CREATE TABLE IF NOT EXISTS meta.table_partitions (
    table_name TEXT NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    partition_name STRING NOT NULL,

    PRIMARY KEY (table_name, partition_name)
);

INSERT INTO meta.table_schemas (table_name, column_name, type)
VALUES
    ('web_requests', 'bytes', 1),
    ('web_requests', 'datetime', 2),
    ('web_requests', 'host', 2),
    ('web_requests', 'method', 2),
    ('web_requests', 'protocol', 2),
    ('web_requests', 'referer', 2),
    ('web_requests', 'status', 2),
    ('web_requests', 'user-identifier', 2),
    ('web_requests', 'timestamp', 2);
CREATE TABLE rss_kv (
    key varchar(100) PRIMARY KEY,
    content text NOT NULL,
    created_at timestamp DEFAULT current_timestamp,
    updated_at timestamp DEFAULT current_timestamp
);

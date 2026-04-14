CREATE TABLE IF NOT EXISTS raw_content (
    url         TEXT        PRIMARY KEY,
    type        TEXT        NOT NULL,
    content     TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_download BOOLEAN     NOT NULL DEFAULT FALSE,
    dismiss     BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_raw_content_type_updated_at
    ON raw_content (type, updated_at DESC);

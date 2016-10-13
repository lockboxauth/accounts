-- +migrate Up
CREATE TABLE accounts (
	id VARCHAR(64) PRIMARY KEY,
	profile_id VARCHAR(36) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	last_used_at TIMESTAMPTZ NOT NULL,
	last_seen_at TIMESTAMPTZ NOT NULL
);

-- +migrate Down
DROP TABLE accounts;

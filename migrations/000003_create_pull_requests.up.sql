BEGIN;

-- Справочник статусов pull request-ов
CREATE TABLE IF NOT EXISTS pr_statuses (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO pr_statuses (name) VALUES ('OPEN'), ('MERGED');

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(500) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    status_id INTEGER NOT NULL REFERENCES pr_statuses(id) DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP NULL
);

CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status_id);
CREATE INDEX idx_pr_created_at ON pull_requests(created_at DESC);

COMMIT;
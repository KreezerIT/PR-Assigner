BEGIN;

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_pr_reviewers_user ON pull_request_reviewers(user_id);
CREATE INDEX idx_pr_reviewers_pr ON pull_request_reviewers(pull_request_id);

COMMIT;
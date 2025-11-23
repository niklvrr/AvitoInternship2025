CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE prs (
    id TEXT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    merged_at TIMESTAMP DEFAULT NULL
);

CREATE INDEX idx_prs_author ON prs(author_id);
CREATE INDEX idx_prs_status ON prs(status);

CREATE TABLE pr_reviewers (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    pr_id TEXT NOT NULL REFERENCES prs(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pr_id)
);

CREATE INDEX idx_pr_reviewers_user ON pr_reviewers(user_id);
CREATE INDEX idx_pr_reviewers_pr ON pr_reviewers(pr_id);
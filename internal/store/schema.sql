CREATE TABLE IF NOT EXISTS runs (
    id              TEXT PRIMARY KEY,
    tag             TEXT NOT NULL,
    workflow        TEXT NOT NULL,
    task_id         TEXT NOT NULL,
    tier            INTEGER NOT NULL,
    task_type       TEXT NOT NULL,
    run_number      INTEGER NOT NULL,
    status          TEXT NOT NULL DEFAULT 'running',
    started_at      TEXT NOT NULL,
    finished_at     TEXT,

    -- correctness
    l1_build        BOOLEAN,
    l2_ut_passed    INTEGER,
    l2_ut_total     INTEGER,
    l3_lint_issues  INTEGER,
    l4_e2e_passed   INTEGER,
    l4_e2e_total    INTEGER,
    correctness_score REAL,

    -- efficiency
    input_tokens    INTEGER,
    output_tokens   INTEGER,
    total_tokens    INTEGER,
    cost_usd        REAL,
    wall_time_secs  REAL,
    tool_uses       INTEGER,

    -- code metrics
    complexity_delta REAL,
    security_issues INTEGER,
    security_fail   BOOLEAN,

    -- rework
    iteration_cycles    INTEGER,
    first_pass_success  BOOLEAN,

    -- LLM Judge
    rubric_correctness  REAL,
    rubric_readability  REAL,
    rubric_simplicity   REAL,
    rubric_robustness   REAL,
    rubric_minimality   REAL,
    rubric_maintainability REAL,
    rubric_go_idioms    REAL,
    rubric_composite    REAL,
    rubric_private_score REAL,

    -- composite
    efficiency_score REAL,
    stability_score REAL,
    final_score     REAL,

    -- environment
    claude_version  TEXT,
    go_version      TEXT,
    judge_model     TEXT,

    -- raw data paths
    raw_json_path   TEXT,
    git_diff_path   TEXT,
    plan_content    TEXT
);

CREATE TABLE IF NOT EXISTS verification_targets (
    run_id      TEXT NOT NULL REFERENCES runs(id),
    vt_id       TEXT NOT NULL,
    passed      BOOLEAN,
    output      TEXT,
    PRIMARY KEY (run_id, vt_id)
);

CREATE TABLE IF NOT EXISTS pairwise_results (
    id              TEXT PRIMARY KEY,
    tag_left        TEXT NOT NULL,
    tag_right       TEXT NOT NULL,
    run_id_left     TEXT REFERENCES runs(id),
    run_id_right    TEXT REFERENCES runs(id),
    task_id         TEXT NOT NULL,
    dimension       TEXT NOT NULL,
    winner          TEXT NOT NULL,
    magnitude       TEXT,
    position_consistent BOOLEAN,
    reasoning       TEXT
);

CREATE INDEX IF NOT EXISTS idx_runs_tag ON runs(tag);
CREATE INDEX IF NOT EXISTS idx_runs_task ON runs(task_id);
CREATE INDEX IF NOT EXISTS idx_runs_status ON runs(status);
CREATE INDEX IF NOT EXISTS idx_vt_run ON verification_targets(run_id);
CREATE INDEX IF NOT EXISTS idx_pw_tags ON pairwise_results(tag_left, tag_right);

CREATE TABLE IF NOT EXISTS pairwise_aggregate (
    id              TEXT PRIMARY KEY,
    tag_left        TEXT NOT NULL,
    tag_right       TEXT NOT NULL,
    task_id         TEXT NOT NULL,
    overall_winner  TEXT NOT NULL,
    weighted_score  REAL,
    confidence      TEXT,
    position_consistent_dims INTEGER,
    total_dims      INTEGER
);

CREATE INDEX IF NOT EXISTS idx_pwa_tags ON pairwise_aggregate(tag_left, tag_right);

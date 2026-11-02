CREATE TABLE IF NOT EXISTS topics (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS questions (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note_id       BIGINT REFERENCES notes(id) ON DELETE SET NULL,
    topic_id      BIGINT REFERENCES topics(id) ON DELETE SET NULL,
    question_text TEXT NOT NULL,
    answer_text   TEXT NOT NULL,
    source        TEXT NOT NULL DEFAULT 'manual' CHECK (source IN ('manual', 'ai')),
    status        TEXT NOT NULL DEFAULT 'approved' CHECK (status IN ('draft', 'approved', 'rejected')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_questions_user_id ON questions(user_id);

CREATE TABLE IF NOT EXISTS review_state (
    question_id      BIGINT PRIMARY KEY REFERENCES questions(id) ON DELETE CASCADE,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ease_factor      NUMERIC(4,2) NOT NULL DEFAULT 2.5,
    interval_days    INT NOT NULL DEFAULT 0,
    repetitions      INT NOT NULL DEFAULT 0,
    due_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_reviewed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_review_state_due ON review_state(user_id, due_at);

CREATE TABLE IF NOT EXISTS answers (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    quality     SMALLINT NOT NULL CHECK (quality BETWEEN 0 AND 5),
    answered_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_answers_user_time ON answers(user_id, answered_at);
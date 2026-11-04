CREATE TABLE IF NOT EXISTS topic_mastery (
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id         BIGINT NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    total_answers    INT NOT NULL DEFAULT 0,
    correct_answers  INT NOT NULL DEFAULT 0,
    last_answered_at TIMESTAMPTZ,
    PRIMARY KEY (user_id, topic_id)
);

CREATE TABLE IF NOT EXISTS processed_events (
    event_id     UUID NOT NULL,
    consumer     TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, consumer)
);
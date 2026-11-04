CREATE TABLE IF NOT EXISTS note_chunks (
    id            BIGSERIAL PRIMARY KEY,
    note_id       BIGINT NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    chunk_index   INT NOT NULL,
    content       TEXT NOT NULL,
    search_vector TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', content)) STORED,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (note_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_note_chunks_search ON note_chunks USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_note_chunks_user ON note_chunks(user_id);
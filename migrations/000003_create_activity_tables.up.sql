-- migrations/000003_create_activity_tables.up.sql

-- Playback Progress Table
CREATE TABLE playback_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    -- Progress stored in MILLISECONDS as a BIGINT
    progress_ms BIGINT NOT NULL DEFAULT 0 CHECK (progress_ms >= 0),
    last_listened_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Composite primary key ensures one progress record per user/track pair
    PRIMARY KEY (user_id, track_id)
);

-- Index for quickly finding recent progress for a user
CREATE INDEX idx_playbackprogress_user_lastlistened ON playback_progress(user_id, last_listened_at DESC);


-- Bookmarks Table
CREATE TABLE bookmarks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    -- Timestamp stored in MILLISECONDS as a BIGINT
    timestamp_ms BIGINT NOT NULL CHECK (timestamp_ms >= 0),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- No updated_at needed if bookmarks are immutable once created (except for deletion)
);

-- Index for efficient listing of bookmarks for a user on a specific track, ordered by time
CREATE INDEX idx_bookmarks_user_track_time ON bookmarks(user_id, track_id, timestamp_ms ASC);
-- Index for listing recent bookmarks for a user across all tracks
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC);

-- No triggers needed here assuming upsert handles last_listened_at
-- and created_at uses default.
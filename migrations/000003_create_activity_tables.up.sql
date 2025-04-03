-- migrations/000003_create_activity_tables.up.sql

-- Playback Progress Table
CREATE TABLE playback_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE,
    -- Progress stored in seconds as an integer
    progress_seconds INTEGER NOT NULL DEFAULT 0 CHECK (progress_seconds >= 0),
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
    -- Timestamp stored in seconds as an integer
    timestamp_seconds INTEGER NOT NULL CHECK (timestamp_seconds >= 0),
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    -- No updated_at needed if bookmarks are immutable once created (except for deletion)
);

-- Index for efficient listing of bookmarks for a user on a specific track, ordered by time
CREATE INDEX idx_bookmarks_user_track_time ON bookmarks(user_id, track_id, timestamp_seconds ASC);
-- Index for listing recent bookmarks for a user across all tracks
CREATE INDEX idx_bookmarks_user_created ON bookmarks(user_id, created_at DESC);


-- Add triggers for playback_progress updated_at (using last_listened_at effectively)
-- We can update last_listened_at on every UPSERT in the repo logic instead of a trigger here.
-- No trigger needed for bookmarks created_at (default value handles it).
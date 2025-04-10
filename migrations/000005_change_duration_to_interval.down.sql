-- migrations/000005_change_duration_to_interval.down.sql

-- Revert changes from the UP migration

-- Change bookmarks.timestamp_ms back to BIGINT
-- Note: EXTRACT(EPOCH FROM ...) gives seconds, multiply by 1000 for milliseconds.
-- Handle potential fractional seconds if necessary, though unlikely with millisecond origin.
ALTER TABLE bookmarks
ALTER COLUMN timestamp_ms TYPE BIGINT
USING (EXTRACT(EPOCH FROM timestamp_ms) * 1000)::bigint;
-- If renamed: ALTER TABLE bookmarks RENAME COLUMN "timestamp" TO timestamp_ms;

-- Change playback_progress.progress_ms back to BIGINT
ALTER TABLE playback_progress
ALTER COLUMN progress_ms TYPE BIGINT
USING (EXTRACT(EPOCH FROM progress_ms) * 1000)::bigint;
-- If renamed: ALTER TABLE playback_progress RENAME COLUMN progress TO progress_ms;

-- Change audio_tracks.duration_ms back to BIGINT
ALTER TABLE audio_tracks
ALTER COLUMN duration_ms TYPE BIGINT
USING (EXTRACT(EPOCH FROM duration_ms) * 1000)::bigint;
-- If renamed: ALTER TABLE audio_tracks RENAME COLUMN duration TO duration_ms;

-- Optional: Restore original CHECK constraints if they were modified
-- Example:
-- ALTER TABLE audio_tracks DROP CONSTRAINT audio_tracks_duration_check;
-- ALTER TABLE audio_tracks ADD CONSTRAINT audio_tracks_duration_ms_check CHECK (duration_ms >= 0);
-- (Do similarly for progress_ms and timestamp_ms if needed)
-- migrations/000005_change_duration_to_interval.up.sql

-- Note: Ensure data types match exactly if there are constraints/indexes depending on them.
-- The USING clause converts existing BIGINT milliseconds to INTERVAL.

-- Change audio_tracks.duration_ms to INTERVAL
ALTER TABLE audio_tracks
ALTER COLUMN duration_ms TYPE INTERVAL
USING (duration_ms * interval '1 millisecond');
-- Optionally rename the column for clarity, but this requires updating Go code references more extensively.
-- Let's keep the name for now to minimize Go code changes, though 'duration' would be better.
-- ALTER TABLE audio_tracks RENAME COLUMN duration_ms TO duration;

-- Change playback_progress.progress_ms to INTERVAL
ALTER TABLE playback_progress
ALTER COLUMN progress_ms TYPE INTERVAL
USING (progress_ms * interval '1 millisecond');
-- ALTER TABLE playback_progress RENAME COLUMN progress_ms TO progress;

-- Change bookmarks.timestamp_ms to INTERVAL
ALTER TABLE bookmarks
ALTER COLUMN timestamp_ms TYPE INTERVAL
USING (timestamp_ms * interval '1 millisecond');
-- ALTER TABLE bookmarks RENAME COLUMN timestamp_ms TO "timestamp"; -- "timestamp" might be a reserved word, quoting is safer

-- Optional: Update CHECK constraints if they were specific to BIGINT ranges
-- Example: (Though INTERVAL naturally handles positive durations)
-- ALTER TABLE audio_tracks DROP CONSTRAINT audio_tracks_duration_ms_check;
-- ALTER TABLE audio_tracks ADD CONSTRAINT audio_tracks_duration_check CHECK (duration_ms >= interval '0 seconds');
-- (Do similarly for progress_ms and timestamp_ms if needed)
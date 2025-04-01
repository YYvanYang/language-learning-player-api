-- migrations/000002_create_audio_tables.down.sql

DROP TRIGGER IF EXISTS update_audio_collections_updated_at ON audio_collections;
DROP TRIGGER IF EXISTS update_audio_tracks_updated_at ON audio_tracks;

DROP TABLE IF EXISTS collection_tracks;
DROP TABLE IF EXISTS audio_collections;
DROP TABLE IF EXISTS audio_tracks;
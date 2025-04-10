-- migrations/000002_create_audio_tables.up.sql

-- Audio Tracks Table
CREATE TABLE audio_tracks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    language_code VARCHAR(10) NOT NULL, -- e.g., 'en-US', 'zh-CN'
    level VARCHAR(50),                  -- e.g., 'A1', 'B2', 'NATIVE' (Matches domain.AudioLevel)
    duration_ms INTERVAL NOT NULL DEFAULT interval '0 seconds' CHECK (duration_ms >= interval '0 seconds'),
    minio_bucket VARCHAR(100) NOT NULL,
    minio_object_key VARCHAR(1024) NOT NULL UNIQUE, -- Object key should be unique within bucket
    cover_image_url VARCHAR(1024),
    uploader_id UUID NULL REFERENCES users(id) ON DELETE SET NULL, -- Optional link to user
    is_public BOOLEAN NOT NULL DEFAULT true,
    tags TEXT[] NULL,                   -- Array of text tags
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for audio_tracks
CREATE INDEX idx_audiotracks_language ON audio_tracks(language_code);
CREATE INDEX idx_audiotracks_level ON audio_tracks(level);
CREATE INDEX idx_audiotracks_uploader ON audio_tracks(uploader_id);
CREATE INDEX idx_audiotracks_is_public ON audio_tracks(is_public);
CREATE INDEX idx_audiotracks_tags ON audio_tracks USING GIN (tags); -- GIN index for array searching
CREATE INDEX idx_audiotracks_created_at ON audio_tracks(created_at DESC);

-- Audio Collections Table
CREATE TABLE audio_collections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Collections deleted if owner is deleted
    type VARCHAR(50) NOT NULL CHECK (type IN ('COURSE', 'PLAYLIST')), -- Matches domain.CollectionType
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for audio_collections
CREATE INDEX idx_audiocollections_owner ON audio_collections(owner_id);
CREATE INDEX idx_audiocollections_type ON audio_collections(type);

-- Collection Tracks Association Table (Many-to-Many with Ordering)
CREATE TABLE collection_tracks (
    collection_id UUID NOT NULL REFERENCES audio_collections(id) ON DELETE CASCADE,
    track_id UUID NOT NULL REFERENCES audio_tracks(id) ON DELETE CASCADE, -- If track deleted, remove from collection
    position INTEGER NOT NULL DEFAULT 0 CHECK (position >= 0), -- Order within the collection
    PRIMARY KEY (collection_id, track_id) -- Ensure a track is only added once per collection
);

-- Index for finding collections a track belongs to
CREATE INDEX idx_collectiontracks_track_id ON collection_tracks(track_id);
-- Index for retrieving tracks in order for a collection
CREATE INDEX idx_collectiontracks_order ON collection_tracks(collection_id, position);


-- Add triggers to automatically update updated_at timestamps for new tables
-- Ensure the function update_updated_at_column exists from migration 000001
CREATE TRIGGER update_audio_tracks_updated_at
BEFORE UPDATE ON audio_tracks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_audio_collections_updated_at
BEFORE UPDATE ON audio_collections
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
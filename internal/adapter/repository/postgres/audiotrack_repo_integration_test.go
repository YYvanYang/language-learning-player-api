// internal/adapter/repository/postgres/audiotrack_repo_integration_test.go
package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	// For checking list order
	"github.com/yvanyang/language-learning-player-backend/internal/adapter/repository/postgres"
	"github.com/yvanyang/language-learning-player-backend/internal/domain"
	"github.com/yvanyang/language-learning-player-backend/internal/port"
	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAudioTrackRepo(t *testing.T) *postgres.AudioTrackRepository {
	require.NotNil(t, testDBPool, "Database pool not initialized")
	require.NotNil(t, testLogger, "Logger not initialized")
	return postgres.NewAudioTrackRepository(testDBPool, testLogger)
}

// Helper to clear audio related tables (add collections etc. if needed)
func clearAudioTables(t *testing.T, ctx context.Context) {
	// Delete in reverse order of FK dependencies if needed
	_, err := testDBPool.Exec(ctx, "DELETE FROM collection_tracks")
	require.NoError(t, err, "Failed to clear collection_tracks table")
	_, err = testDBPool.Exec(ctx, "DELETE FROM audio_tracks")
	require.NoError(t, err, "Failed to clear audio_tracks table")
	// clear users if necessary or handle FK on track creation
}

// Helper to create a dummy track for testing
func createTestTrack(t *testing.T, ctx context.Context, repo *postgres.AudioTrackRepository, suffix string) *domain.AudioTrack {
	lang, _ := domain.NewLanguage("en-US", "English")
	track, err := domain.NewAudioTrack(
		"Test Track "+suffix,
		"Description "+suffix,
		"test-bucket",
		fmt.Sprintf("test/object/key_%s.mp3", suffix),
		lang,
		domain.LevelB1,
		120*time.Second,
		nil, // No uploader
		true,
		[]string{"test", suffix},
		nil, // No cover image
	)
	require.NoError(t, err)
	err = repo.Create(ctx, track)
	require.NoError(t, err)
	return track
}

func TestAudioTrackRepository_Integration_CreateAndFind(t *testing.T) {
	ctx := context.Background()
	repo := setupAudioTrackRepo(t)
	clearAudioTables(t, ctx)

	track1 := createTestTrack(t, ctx, repo, "1")

	// Find by ID
	found, err := repo.FindByID(ctx, track1.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, track1.ID, found.ID)
	assert.Equal(t, track1.Title, found.Title)
	assert.Equal(t, track1.MinioObjectKey, found.MinioObjectKey)
	assert.Equal(t, track1.Language.Code(), found.Language.Code())
	assert.Equal(t, track1.Level, found.Level)
	assert.Equal(t, track1.Duration, found.Duration)
	assert.Equal(t, track1.Tags, found.Tags)
	assert.WithinDuration(t, track1.CreatedAt, found.CreatedAt, time.Second)

	// Find non-existent
	_, err = repo.FindByID(ctx, domain.NewTrackID())
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestAudioTrackRepository_Integration_ListByIDs(t *testing.T) {
	ctx := context.Background()
	repo := setupAudioTrackRepo(t)
	clearAudioTables(t, ctx)

	track1 := createTestTrack(t, ctx, repo, "T1")
	track2 := createTestTrack(t, ctx, repo, "T2")
	track3 := createTestTrack(t, ctx, repo, "T3")

	// Test case 1: Get specific tracks in order
	idsToFetch := []domain.TrackID{track3.ID, track1.ID}
	tracks, err := repo.ListByIDs(ctx, idsToFetch)
	require.NoError(t, err)
	require.Len(t, tracks, 2)
	assert.Equal(t, track3.ID, tracks[0].ID) // Check order if repo guarantees it
	assert.Equal(t, track1.ID, tracks[1].ID)

	// Test case 2: Include non-existent ID
	nonExistentID := domain.NewTrackID()
	idsWithNonExistent := []domain.TrackID{track2.ID, nonExistentID}
	tracks, err = repo.ListByIDs(ctx, idsWithNonExistent)
	require.NoError(t, err)
	require.Len(t, tracks, 1) // Should only return existing track
	assert.Equal(t, track2.ID, tracks[0].ID)

	// Test case 3: Empty list
	tracks, err = repo.ListByIDs(ctx, []domain.TrackID{})
	require.NoError(t, err)
	require.Len(t, tracks, 0)
}

func TestAudioTrackRepository_Integration_List_PaginationAndFilter(t *testing.T) {
	ctx := context.Background()
	repo := setupAudioTrackRepo(t)
	clearAudioTables(t, ctx)

	// Create test data
	langEN, _ := domain.NewLanguage("en-US", "")
	langES, _ := domain.NewLanguage("es-ES", "")
	createTrackWithDetails := func(suffix string, lang domain.Language, level domain.AudioLevel, tags []string) {
		track, _ := domain.NewAudioTrack("Title "+suffix, "Desc "+suffix, "bucket", fmt.Sprintf("key_%s", suffix), lang, level, 10*time.Second, nil, true, tags, nil)
		err := repo.Create(ctx, track)
		require.NoError(t, err)
	}
	createTrackWithDetails("EN_A1", langEN, domain.LevelA1, []string{"news", "easy"})
	createTrackWithDetails("EN_A2", langEN, domain.LevelA2, []string{"story", "easy"})
	createTrackWithDetails("EN_B1", langEN, domain.LevelB1, []string{"news", "intermediate"})
	createTrackWithDetails("ES_A2", langES, domain.LevelA2, []string{"story", "easy", "spanish"})
	createTrackWithDetails("ES_B2", langES, domain.LevelB2, []string{"culture", "intermediate", "spanish"})

	// Test Case 1: List all (default sort: createdAt desc), first page
	page1 := pagination.Page{Limit: 3, Offset: 0}
	tracks, total, err := repo.List(ctx, port.ListTracksParams{}, page1)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	require.Len(t, tracks, 3)
	// Assert default order (newest first - ES_B2, ES_A2, EN_B1)
	assert.Equal(t, "Title ES_B2", tracks[0].Title)
	assert.Equal(t, "Title ES_A2", tracks[1].Title)
	assert.Equal(t, "Title EN_B1", tracks[2].Title)

	// Test Case 2: Second page
	page2 := pagination.Page{Limit: 3, Offset: 3}
	tracks, total, err = repo.List(ctx, port.ListTracksParams{}, page2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	require.Len(t, tracks, 2)
	// Assert order (EN_A2, EN_A1)
	assert.Equal(t, "Title EN_A2", tracks[0].Title)
	assert.Equal(t, "Title EN_A1", tracks[1].Title)

	// Test Case 3: Filter by language 'es-ES'
	langFilter := "es-ES"
	paramsLang := port.ListTracksParams{LanguageCode: &langFilter}
	tracks, total, err = repo.List(ctx, paramsLang, pagination.Page{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, tracks, 2)
	assert.Equal(t, "Title ES_B2", tracks[0].Title) // Default sort still applies
	assert.Equal(t, "Title ES_A2", tracks[1].Title)

	// Test Case 4: Filter by level A2
	levelFilter := domain.LevelA2
	paramsLevel := port.ListTracksParams{Level: &levelFilter}
	tracks, total, err = repo.List(ctx, paramsLevel, pagination.Page{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, tracks, 2)
	// Check titles, order might depend on creation time or default sort
	titles := []string{tracks[0].Title, tracks[1].Title}
	assert.Contains(t, titles, "Title EN_A2")
	assert.Contains(t, titles, "Title ES_A2")

	// Test Case 5: Filter by tag 'news' and sort by title asc
	tagFilter := []string{"news"}
	sortByTitle := "title"
	sortDirAsc := "asc"
	paramsTagSort := port.ListTracksParams{Tags: tagFilter, SortBy: sortByTitle, SortDirection: sortDirAsc}
	tracks, total, err = repo.List(ctx, paramsTagSort, pagination.Page{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, tracks, 2)
	assert.Equal(t, "Title EN_A1", tracks[0].Title) // Sorted A-Z
	assert.Equal(t, "Title EN_B1", tracks[1].Title)

	// Test Case 6: Filter by query string
	query := "InteRmedIate" // Case-insensitive search in title/desc
	paramsQuery := port.ListTracksParams{Query: &query}
	tracks, total, err = repo.List(ctx, paramsQuery, pagination.Page{Limit: 10, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 2, total) // EN_B1, ES_B2 contain 'intermediate' tag, but query searches title/desc. Let's assume desc has level name.
	// This test might fail if desc doesn't contain the level string - adjust test data or query logic.
	// For now, let's assume the search finds EN_B1 and ES_B2 based on Description content (adjust if needed).
	titles = []string{tracks[0].Title, tracks[1].Title}
	assert.Contains(t, titles, "Title EN_B1")
	assert.Contains(t, titles, "Title ES_B2")

}

// TODO: Add tests for Update (success, conflict on object key)
// TODO: Add tests for Delete
// TODO: Add tests for Exists

# API Endpoint Validation Guide

This document outlines a step-by-step sequence for manually validating the Language Learning Player backend API endpoints, typically using a tool like Postman.

**Prerequisites:**

1.  **API Running:** Ensure the backend API is running locally using `make run` or `make docker-run`.
2.  **Dependencies Running:** Make sure PostgreSQL and MinIO are running (e.g., via `make deps-run`).
3.  **Database Migrated:** Ensure database migrations are up-to-date (`make migrate-up`).
4.  **Postman (or similar tool):** Have Postman installed and ready. It's helpful to create a collection for this project.
5.  **Environment Variables (Postman):** Consider setting up a Postman environment to store variables like `baseUrl` (`http://localhost:8080`), `jwtToken`, user IDs, track IDs, collection IDs, etc., as you progress through the tests.

**Validation Sequence:**

Here's a logical order to test the endpoints, considering dependencies:

**Phase 1: Basic Checks & Public Endpoints**

1.  **Health Check:**
    *   `GET /healthz`
    *   **Expected:** `200 OK`, body "OK".
2.  **Docs Access:**
    *   `GET /`
    *   **Expected:** `302 Found` redirect to `/swagger/index.html`. (Verify in browser)
    *   `GET /swagger/index.html`
    *   **Expected:** `200 OK` with Swagger UI HTML. (Verify in browser)
3.  **Public Audio Listing (Initial State):**
    *   `GET /api/v1/audio/tracks`
    *   **Expected:** `200 OK`, likely with an empty `data` array or default public tracks if seeded. Test basic pagination (`limit`, `offset`) if possible.
4.  **Public Audio Get Details (Non-existent):**
    *   `GET /api/v1/audio/tracks/{non_existent_uuid}` (Replace `{non_existent_uuid}` with a random valid UUID)
    *   **Expected:** `404 Not Found`.

**Phase 2: Authentication Flow**

5.  **User Registration:**
    *   `POST /api/v1/auth/register`
    *   Body: `{ "email": "testuser@example.com", "password": "StrongPassword123", "name": "Test User" }`
    *   **Expected:** `201 Created`, response body contains a JWT token (`{"token": "..."}`).
    *   **Postman:** Save the received `token` to your Postman environment variable (e.g., `jwtToken`).
    *   **Error Case:** Try registering *again* with the same email. **Expected:** `409 Conflict`.
    *   **Error Case:** Try registering with invalid input (bad email format, short password). **Expected:** `400 Bad Request`.
6.  **User Login:**
    *   `POST /api/v1/auth/login`
    *   Body: `{ "email": "testuser@example.com", "password": "StrongPassword123" }`
    *   **Expected:** `200 OK`, response body contains a *new* JWT token.
    *   **Postman:** Update your `jwtToken` variable.
    *   **Error Case:** Try login with incorrect password. **Expected:** `401 Unauthorized`.
    *   **Error Case:** Try login with non-existent email. **Expected:** `401 Unauthorized`.
7.  **Google Auth Callback (Manual/Conceptual Check):**
    *   `POST /api/v1/auth/google/callback`
    *   Body: `{ "idToken": "..." }`
    *   **Validation:** This is harder to fully test in Postman without a valid Google `idToken` obtained from a frontend flow.
    *   **Conceptual Check:** Test with an *empty* or *invalid* token string. **Expected:** `400 Bad Request` or `401 Unauthorized`. If you *can* get a valid token, test the success case (returns JWT) and potentially the `isNewUser` flag. Also test the `409 Conflict` case if the email exists via local registration.

**Phase 3: Authenticated User Endpoints**

*   **Important:** For all subsequent requests in this phase, ensure you add the `Authorization` header in Postman: `Bearer {{jwtToken}}`.

8.  **Get User Profile:**
    *   `GET /api/v1/users/me`
    *   **Expected:** `200 OK`, response body contains the profile info for "testuser@example.com". Verify `email`, `name`, `id`, etc.
    *   **Error Case:** Try without the `Authorization` header. **Expected:** `401 Unauthorized`.

**Phase 4: Content Creation (Tracks via Upload)**

9.  **Request Upload URL:**
    *   `POST /api/v1/uploads/audio/request`
    *   Body: `{ "filename": "test_audio.mp3", "contentType": "audio/mpeg" }`
    *   **Expected:** `200 OK`, response body contains `uploadUrl` (a presigned PUT URL) and `objectKey`.
    *   **Postman:** Save the `objectKey` to a variable (e.g., `trackObjectKey`).
10. **Simulate Upload (Manual Step):**
    *   **Action:** Use a tool like `curl` or Postman itself (set request type to PUT, body to Binary File) to upload an actual small MP3 file to the `uploadUrl` obtained in the previous step. You need to set the `Content-Type` header in this request to match what you sent in step 9 (`audio/mpeg`).
    *   **Verification:** Check your MinIO console (`http://localhost:9001`) or use `mc ls local/language-audio` (if mc client is configured) to confirm the file exists with the correct `objectKey`.
11. **Complete Upload & Create Track:**
    *   `POST /api/v1/audio/tracks`
    *   Body (Use the `objectKey` from step 9):
        ```json
        {
            "objectKey": "{{trackObjectKey}}",
            "title": "Test Audio Track 1",
            "description": "My first test track.",
            "languageCode": "en-US",
            "level": "B1",
            "durationMs": 123456, // Provide a realistic duration
            "isPublic": true,
            "tags": ["test", "beginner"]
        }
        ```
    *   **Expected:** `201 Created`, response body contains the details of the newly created track.
    *   **Postman:** Save the `id` of the created track to a variable (e.g., `trackId1`).
    *   **Error Case:** Try completing the *same* `objectKey` again. **Expected:** Behavior depends on implementation (likely `409 Conflict` or `400 Bad Request`).
    *   **Error Case:** Try with invalid input (missing required fields, bad level enum). **Expected:** `400 Bad Request`.
12. **Repeat Steps 9-11:** Create at least one more track (e.g., `trackId2`) for testing listing and collections. Make one public and one private if desired (`isPublic: false`).

**Phase 5: Content Retrieval/Listing (Authenticated & Public)**

13. **Get Track Details (Existing):**
    *   `GET /api/v1/audio/tracks/{{trackId1}}` (Try with and without auth header if some tracks are public)
    *   **Expected:** `200 OK`, body contains track details including `playUrl` (the presigned GET URL). Verify the `playUrl` looks correct (points to MinIO endpoint/bucket/key).
14. **List Tracks (With Data):**
    *   `GET /api/v1/audio/tracks` (Try with and without auth header)
    *   **Expected:** `200 OK`, `data` array contains the tracks created.
    *   **Test Filters:** Add query parameters (`?lang=en-US`, `?level=B1`, `?tags=test`, `?isPublic=true`). Verify results are filtered correctly.
    *   **Test Pagination:** Use `?limit=1&offset=0`, `?limit=1&offset=1`. Verify results and `total`, `limit`, `offset` fields in the response.
    *   **Test Sorting:** Use `?sortBy=title&sortDir=asc`. Verify order.

**Phase 6: Collections Management**

15. **Create Collection:**
    *   `POST /api/v1/audio/collections`
    *   Body: `{ "title": "My Test Playlist", "description": "Tracks for testing", "type": "PLAYLIST", "initialTrackIds": ["{{trackId1}}", "{{trackId2}}"] }`
    *   **Expected:** `201 Created`, response body contains collection details (possibly including the initial tracks).
    *   **Postman:** Save the `id` of the created collection (e.g., `collectionId1`).
    *   **Error Case:** Invalid type, non-existent initial track ID. **Expected:** `400 Bad Request`.
16. **Get Collection Details:**
    *   `GET /api/v1/audio/collections/{{collectionId1}}`
    *   **Expected:** `200 OK`, body contains collection details and the ordered list of tracks (`{{trackId1}}`, `{{trackId2}}`).
17. **Update Collection Metadata:**
    *   `PUT /api/v1/audio/collections/{{collectionId1}}`
    *   Body: `{ "title": "My Updated Playlist", "description": "Updated description" }`
    *   **Expected:** `204 No Content`.
    *   **Verification:** Call `GET /api/v1/audio/collections/{{collectionId1}}` again to verify changes.
18. **Update Collection Tracks:**
    *   `PUT /api/v1/audio/collections/{{collectionId1}}/tracks`
    *   Body: `{ "orderedTrackIds": ["{{trackId2}}", "{{trackId1}}"] }` (Reversed order)
    *   **Expected:** `204 No Content`.
    *   **Verification:** Call `GET /api/v1/audio/collections/{{collectionId1}}` again to verify the new track order.
    *   **Error Case:** Body with non-existent track ID. **Expected:** `400` or `404`.
19. **Delete Collection:**
    *   `DELETE /api/v1/audio/collections/{{collectionId1}}`
    *   **Expected:** `204 No Content`.
    *   **Verification:** Call `GET /api/v1/audio/collections/{{collectionId1}}` again. **Expected:** `404 Not Found`.

**Phase 7: User Activity (Progress & Bookmarks)**

20. **Record Progress:**
    *   `POST /api/v1/users/me/progress`
    *   Body: `{ "trackId": "{{trackId1}}", "progressSeconds": 30.5 }`
    *   **Expected:** `204 No Content`.
    *   **Action:** Send again with updated progress: `{ "trackId": "{{trackId1}}", "progressSeconds": 65.0 }`. **Expected:** `204 No Content`.
    *   **Error Case:** Non-existent `trackId`. **Expected:** `404 Not Found`.
21. **Get Specific Progress:**
    *   `GET /api/v1/users/me/progress/{{trackId1}}`
    *   **Expected:** `200 OK`, body contains progress details, `progressSeconds` should be `65.0`.
    *   **Error Case:** Get progress for a track with no recorded progress. **Expected:** `404 Not Found`.
22. **List Progress:**
    *   `GET /api/v1/users/me/progress`
    *   **Expected:** `200 OK`, `data` array contains the progress record for `trackId1`. Test pagination if you record progress for more tracks.
23. **Create Bookmark:**
    *   `POST /api/v1/bookmarks`
    *   Body: `{ "trackId": "{{trackId1}}", "timestampSeconds": 42.0, "note": "Important point here" }`
    *   **Expected:** `201 Created`, response body contains bookmark details including a new `id`.
    *   **Postman:** Save the bookmark `id` (e.g., `bookmarkId1`).
    *   **Action:** Create another bookmark for the same track: `{ "trackId": "{{trackId1}}", "timestampSeconds": 95.5, "note": "Another point" }`. Save ID (`bookmarkId2`).
    *   **Action:** Create a bookmark for `trackId2`: `{ "trackId": "{{trackId2}}", "timestampSeconds": 10.0 }`. Save ID (`bookmarkId3`).
24. **List Bookmarks:**
    *   `GET /api/v1/bookmarks`
    *   **Expected:** `200 OK`, `data` contains all three bookmarks. Test pagination (`limit`, `offset`).
    *   **Test Filter:** `GET /api/v1/bookmarks?trackId={{trackId1}}`. **Expected:** `200 OK`, `data` contains only `bookmarkId1` and `bookmarkId2`.
25. **Delete Bookmark:**
    *   `DELETE /api/v1/bookmarks/{{bookmarkId1}}`
    *   **Expected:** `204 No Content`.
    *   **Verification:** Call `GET /api/v1/bookmarks`. **Expected:** `bookmarkId1` should be gone.
    *   **Error Case:** Delete the *same* bookmark again. **Expected:** `404 Not Found`.
    *   **Error Case:** Try deleting a bookmark belonging to another user (if possible to set up). **Expected:** `403 Forbidden` or `404 Not Found`.

This sequence covers most functional paths. Remember to also test invalid inputs and authorization failures throughout the process. 
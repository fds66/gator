-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id,last_fetched_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;
-- name: ListFeeds :many
SELECT *
FROM feeds;
-- name: FeedFromID :one
SELECT *
FROM feeds
WHERE id = $1;
-- name: FeedFromURL :one
SELECT *
FROM feeds
WHERE url = $1;
-- name: ResetFeeds :exec
DELETE FROM feeds;
-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1, updated_at = $2
WHERE id = $3;
-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
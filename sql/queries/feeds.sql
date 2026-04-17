-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: GetFeedByURL :one
SELECT *
FROM feeds
WHERE url = $1;
-- name: GetFeeds :many
SELECT feeds.name,
    feeds.url,
    users.name AS username
FROM feeds
    LEFT JOIN users ON users.id = feeds.user_id;
-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING *
)
SELECT inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
    INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
    INNER JOIN users ON users.id = inserted_feed_follow.user_id;
-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM feed_follows
    INNER JOIN feeds ON feeds.id = feed_follows.feed_id
    INNER JOIN users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1;
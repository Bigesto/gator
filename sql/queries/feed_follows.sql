-- name: CreateFeedFollows :one
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1::UUID,
        $2::TIMESTAMP,
        $3::TIMESTAMP,
        $4::UUID,
        $5::UUID
    )
    RETURNING id, created_at, updated_at, user_id, feed_id
) SELECT inserted_feed_follows.id,
    inserted_feed_follows.created_at,
    inserted_feed_follows.updated_at,
    inserted_feed_follows.user_id,
    inserted_feed_follows.feed_id,
users.name AS user_name,
feeds.name AS feed_name
FROM inserted_feed_follows
INNER JOIN users ON inserted_feed_follows.user_id = users.id
INNER JOIN feeds ON inserted_feed_follows.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
-- GetFeedFollowsForUser: (user_id uuid.UUID) --
SELECT
    feed_follows.id,
    feed_follows.created_at,
    feed_follows.updated_at,
    feed_follows.user_id,
    feed_follows.feed_id,
    users.name AS user_name,
    feeds.name AS feed_name
FROM feed_follows
INNER JOIN users ON feed_follows.user_id = users.id
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = sqlc.arg('user_id')::uuid;

-- name: DeleteFeedFollowsByUserAndURL :one
DELETE FROM feed_follows
WHERE user_id = $1
  AND feed_id = (SELECT id FROM feeds WHERE url = $2)
RETURNING feed_id;
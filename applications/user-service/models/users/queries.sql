-- name: Attributes :one
-- Attributes will use the user's [User.ID] to hydrate all available [User] attribute(s). Note that the following call is more taxing on the database.
SELECT *
FROM "User"
WHERE (id) = sqlc.arg(id)::bigserial
  AND (deletion) IS NULL
LIMIT 1;

-- name: Create :one
-- Create will create a new [User] record.
INSERT INTO "User" (email) VALUES ($1) RETURNING *;

-- name: UpdateUserAvatar :exec
-- UpdateUserAvatar will update a provided [User] with their specified avatar.
UPDATE "User" SET avatar = $2, modification = $3 WHERE (email) = $1 AND (deletion) IS NULL;

-- name: Me :one
-- Me will return a [User] and all associated attribute(s) when provided the User's email address.
SELECT * FROM "User" WHERE email = $1;

-- name: Count :one
-- Count returns 0 or 1 depending on if a [User] record matching the provided email exists.
SELECT count(*) FROM "User" WHERE (email) = sqlc.arg(email)::text AND (deletion) IS NULL;

-- name: Total :one
-- Total returns the total number of [User] records, excluding deleted record(s).
SELECT count(*) FROM "User" WHERE (deletion) IS NULL;

-- name: All :one
-- All returns the total number of [User] records, including deleted record(s).
SELECT count(*) FROM "User";

-- name: List :many
-- List returns all active User record(s).
SELECT * FROM "User" WHERE (deletion) IS NULL;

-- name: Users :many
-- Users returns all User record(s).
SELECT * FROM "User";

-- name: Delete :exec
-- Delete performs a hard database delete on a [User] record.
DELETE FROM "User" WHERE id = $1;

-- name: DeleteByEmail :exec
-- DeleteByEmail performs a hard database delete on a [User] record.
DELETE FROM "User" WHERE email = $1;

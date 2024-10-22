-- name: Create :one
-- Create creates a new [User] database record.
INSERT INTO "User" (email, password) VALUES ($1, $2) RETURNING *;

-- name: Get :one
SELECT id, email, password FROM "User" WHERE email = $1 AND (deletion) IS NULL;

-- name: Count :one
-- Count returns 0 or 1 depending on if a User record matching the provided email exists.
SELECT count(*) FROM "User" WHERE (email) = sqlc.arg(email)::text AND (deletion) IS NULL;

-- name: Clean :exec
-- Clean performs a hard delete on the [User] database record, regardless if a soft delete has been performed, and only by email. This function should only be used in test(s).
DELETE FROM "User" WHERE (email) = (sqlc.arg(email)::text);

-- name: Delete :exec
-- Delete performs a hard delete on the [User] database record, regardless if a soft delete has been performed.
DELETE FROM "User" WHERE (id, email) = (sqlc.arg(id), sqlc.arg(email)::text);

-- name: Remove :exec
-- Remove performs a soft delete on the [User] database record.
UPDATE "User" SET (modification, deletion) = (now(), now()) WHERE (id, email) = (sqlc.arg(id), sqlc.arg(email)::text) AND (deletion) IS NULL;

-- name: Extract :one
-- Extract retrieves a given [User] database record, regardless of its deletion status.
SELECT * FROM "User" WHERE (id, email) = (sqlc.arg(id), sqlc.arg(email)::text);

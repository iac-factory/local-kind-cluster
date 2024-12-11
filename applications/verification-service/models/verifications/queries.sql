-- name: Create :one
-- Create establishes a new [Verification] database record.
INSERT INTO "Verification" (email, code) VALUES ($1, $2) RETURNING *;

-- name: Count :one
-- Count returns 0 or 1 depending on if a Verification record matching the provided email exists.
SELECT count(*) FROM "Verification" WHERE (email) = sqlc.arg(email)::text AND (deletion) IS NULL;

-- name: Get :one
-- Get returns a fully hydrated [Verification] database record if a match is found via email.
SELECT * FROM "Verification" WHERE (email) = sqlc.arg(email)::text AND (deletion) IS NULL;

-- name: Status :one
-- Status returns a partially hydrated [Verification] database record only including the user's email and verified attribute(s).
SELECT email, verified FROM "Verification" WHERE (email) = $1 AND (deletion) IS NULL;

-- name: Verify :exec
-- Verify updates the [Verification] database record with a verified state.
UPDATE "Verification" SET verified = true, modification = $2 WHERE (email) = $1 AND (deletion) IS NULL;

-- name: Delete :exec
-- Delete performs a hard database delete on a [Verification] record.
DELETE FROM "Verification" WHERE id = $1;

-- name: DeleteByEmail :exec
-- DeleteByEmail performs a hard database delete on a [Verification] record.
DELETE FROM "Verification" WHERE email = $1;

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

-- name: ExistsForce :one
-- Exists checks if a [User] record exists, searching for the entry via the [User.ID] property, regardless if a user has been soft deleted.
SELECT EXISTS (SELECT 1 FROM "User" WHERE (id) = sqlc.arg(id));

-- name: Exists :one
-- Exists checks if a [User] record exists, searching for the entry via the [User.ID] property.
SELECT EXISTS (SELECT 1 FROM "User" WHERE (id) = sqlc.arg(id) AND (deletion) IS NULL);

-- name: UpdateUserAvatar :exec
-- UpdateUserAvatar will update a provided [User] with their specified avatar.
UPDATE "User" SET avatar = sqlc.arg(avatar)::text, modification = now() WHERE (id) = sqlc.arg(id) AND (deletion) IS NULL;

-- name: Me :one
-- Me will return a [User] and all associated attribute(s) when provided the User's email address.
SELECT * FROM "User" WHERE email = $1;

-- name: GetUserEmailAddressByID :one
-- GetUserEmailAddressByID will return a [User] with the record's [User.Email] and [User.ID] hydrated when searching by a [User] identifier.
SELECT "id", "email" FROM "User" WHERE id = sqlc.arg(id) AND (deletion) IS NULL;

-- name: GetUserEmailAddressByIDForce :one
-- GetUserEmailAddressByIDForce will return a [User] with the record's [User.Email] and [User.ID] hydrated when searching by a [User] identifier -- regardless of soft delete.
SELECT "id", "email" FROM "User" WHERE id = sqlc.arg(id);

-- name: Count :one
-- Count returns 0 or 1 depending on if a [User] record matching the provided email exists.
SELECT count(*) FROM "User" WHERE (email) = sqlc.arg(email) AND (deletion) IS NULL;

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

-- name: Clean :exec
-- Clean performs a hard delete on the [User] database record, regardless if a soft delete has been performed, and only by email. This function should only be used in test(s).
DELETE FROM "User" WHERE (email) = sqlc.arg(email);

-- name: DeleteHard :exec
-- DeleteHard performs a hard delete on the [User] database record, regardless if a soft delete has been performed.
DELETE FROM "User" WHERE (id) = sqlc.arg(id);

-- name: DeleteSoft :exec
-- DeleteSoft performs a soft delete on the [User] database record if the record hasn't already been deleted.
UPDATE "User" SET (modification, deletion) = (now(), now()) WHERE (id) = (sqlc.arg(id)) AND (deletion) IS NULL;

-- name: Extract :one
-- Extract retrieves a given [User] database record, regardless of its deletion status.
SELECT * FROM "User" WHERE (id, email) = (sqlc.arg(id), sqlc.arg(email));

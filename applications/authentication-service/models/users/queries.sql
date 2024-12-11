-- name: Create :one
-- Create creates a new [User] database record.
INSERT INTO "User" (email, password) VALUES ($1, $2) RETURNING *;

-- name: Get :one
SELECT id, email, password FROM "User" WHERE email = $1 AND (deletion) IS NULL;

-- name: GetForce :one
SELECT id, email, password FROM "User" WHERE email = $1;

-- name: ExistsForce :one
-- Exists checks if a [User] record exists, searching for the entry via the [User.ID] property, regardless if a user has been soft deleted.
SELECT EXISTS (SELECT 1 FROM "User" WHERE (id) = sqlc.arg(id));

-- name: Exists :one
-- Exists checks if a [User] record exists, searching for the entry via the [User.ID] property.
SELECT EXISTS (SELECT 1 FROM "User" WHERE (id) = sqlc.arg(id) AND (deletion) IS NULL);

-- name: GetUserEmailAddressByID :one
-- GetUserEmailAddressByID will return a [User] with the record's [User.Email] and [User.ID] hydrated when searching by a [User] identifier.
SELECT "id", "email" FROM "User" WHERE id = sqlc.arg(id) AND (deletion) IS NULL;

-- name: GetUserEmailAddressByIDForce :one
-- GetUserEmailAddressByIDForce will return a [User] with the record's [User.Email] and [User.ID] hydrated when searching by a [User] identifier -- regardless of soft delete.
SELECT "id", "email" FROM "User" WHERE id = sqlc.arg(id);

-- name: Count :one
-- Count returns 0 or 1 depending on if a User record matching the provided email exists.
SELECT count(*) FROM "User" WHERE (email) = sqlc.arg(email)::text AND (deletion) IS NULL;

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

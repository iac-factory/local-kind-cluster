// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package users

import (
	"context"
)

type Querier interface {
	// Clean performs a hard delete on the [User] database record, regardless if a soft delete has been performed, and only by email. This function should only be used in test(s).
	Clean(ctx context.Context, db DBTX, email string) error
	// Count returns 0 or 1 depending on if a User record matching the provided email exists.
	Count(ctx context.Context, db DBTX, email string) (int64, error)
	// Create creates a new [User] database record.
	Create(ctx context.Context, db DBTX, arg *CreateParams) (User, error)
	// Delete performs a hard delete on the [User] database record, regardless if a soft delete has been performed.
	Delete(ctx context.Context, db DBTX, arg *DeleteParams) error
	// Extract retrieves a given [User] database record, regardless of its deletion status.
	Extract(ctx context.Context, db DBTX, arg *ExtractParams) (User, error)
	Get(ctx context.Context, db DBTX, email string) (GetRow, error)
	// Remove performs a soft delete on the [User] database record.
	Remove(ctx context.Context, db DBTX, arg *RemoveParams) error
}

var _ Querier = (*Queries)(nil)

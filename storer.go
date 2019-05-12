package accounts

import "context"

// Storer dictates how Accounts will be persisted and how to
// interact with those persisted Accounts.
type Storer interface {
	Create(ctx context.Context, account Account) error
	Get(ctx context.Context, id string) (Account, error)
	Update(ctx context.Context, id string, change Change) error
	Delete(ctx context.Context, id string) error
	ListByProfile(ctx context.Context, profileID string) ([]Account, error)
}

package memory

import (
	"context"

	"lockbox.dev/accounts"
)

// Factory is a generator of Storers for testing purposes.
type Factory struct{}

// NewStorer creates a new, isolated, in-memory Storer for tests.
func (Factory) NewStorer(_ context.Context) (accounts.Storer, error) { //nolint:ireturn // interface requires returning an interface
	return NewStorer()
}

// TeardownStorers does nothing and is only included to fill an interface.
func (Factory) TeardownStorers() error {
	return nil
}

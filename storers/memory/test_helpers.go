package memory

import (
	"context"

	"impractical.co/auth/accounts"
)

type Factory struct{}

func (m Factory) NewStorer(ctx context.Context) (accounts.Storer, error) {
	return NewStorer()
}

func (m Factory) TeardownStorers() error {
	return nil
}

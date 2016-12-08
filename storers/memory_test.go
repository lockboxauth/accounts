package storers

import (
	"context"

	"code.impractical.co/accounts"
)

func init() {
	storerFactories = append(storerFactories, MemstoreFactory{})
}

type MemstoreFactory struct{}

func (m MemstoreFactory) NewStorer(ctx context.Context) (accounts.Storer, error) {
	return NewMemstore()
}

func (m MemstoreFactory) TeardownStorers() error {
	return nil
}

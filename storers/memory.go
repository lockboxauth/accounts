package storers

import (
	"context"
	"sync"

	"code.impractical.co/accounts"
)

// Memstore is an in-memory implementation of the Storer
// interface.
type Memstore struct {
	accounts map[string]accounts.Account
	lock     sync.RWMutex
}

// NewMemstore returns a Memstore instance that is ready
// to be used as a Storer.
func NewMemstore() *Memstore {
	return &Memstore{
		accounts: map[string]accounts.Account{},
	}
}

// Create inserts the passed Account into the Memstore,
// returning an ErrAccountAlreadyExists error if an Account
// with the same ID already exists in the Memstore.
func (m *Memstore) Create(ctx context.Context, account accounts.Account) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.accounts[account.ID]; ok {
		return accounts.ErrAccountAlreadyExists
	}
	m.accounts[account.ID] = account
	return nil
}

// Get retrieves the Account specified by the passed ID from
// the Memstore, returning an ErrAccountNotFound error if no
// Account matches the passed ID.
func (m *Memstore) Get(ctx context.Context, id string) (accounts.Account, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	account, ok := m.accounts[id]
	if !ok {
		return accounts.Account{}, accounts.ErrAccountNotFound
	}

	return account, nil
}

// Update applies the passed Change to the Account that matches
// the specified ID in the Memstore, if any Account matches the
// specified ID in the Memstore.
func (m *Memstore) Update(ctx context.Context, id string, change accounts.Change) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	account, ok := m.accounts[id]
	if !ok {
		return nil
	}
	updated := accounts.Apply(change, account)
	m.accounts[id] = updated
	return nil
}

// Delete removes the Account that matches the specified ID from
// the Memstore, if any Account matches the specified ID in the
// Memstore.
func (m *Memstore) Delete(ctx context.Context, id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.accounts, id)
	return nil
}

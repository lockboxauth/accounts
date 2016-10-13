package storers

import (
	"context"
	"sync"

	"code.impractical.co/accounts"
)

type Memstore struct {
	accounts map[string]accounts.Account
	lock     sync.RWMutex
}

func NewMemstore() *Memstore {
	return &Memstore{
		accounts: map[string]accounts.Account{},
	}
}

func (m *Memstore) Create(ctx context.Context, account accounts.Account) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.accounts[account.ID]; ok {
		return accounts.ErrAccountAlreadyExists
	}
	m.accounts[account.ID] = account
	return nil
}

func (m *Memstore) Get(ctx context.Context, id string) (accounts.Account, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	account, ok := m.accounts[id]
	if !ok {
		return accounts.Account{}, accounts.ErrAccountNotFound
	}

	return account, nil
}

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

func (m *Memstore) Delete(ctx context.Context, id string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.accounts, id)
	return nil
}

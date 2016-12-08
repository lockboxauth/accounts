package storers

import (
	"context"

	memdb "github.com/hashicorp/go-memdb"

	"code.impractical.co/accounts"
)

var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": &memdb.TableSchema{
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID"},
					},
					"profileID": &memdb.IndexSchema{
						Name:    "profileID",
						Indexer: &memdb.StringFieldIndex{Field: "ProfileID"},
					},
				},
			},
		},
	}
)

// Memstore is an in-memory implementation of the Storer
// interface.
type Memstore struct {
	db *memdb.MemDB
}

// NewMemstore returns a Memstore instance that is ready
// to be used as a Storer.
func NewMemstore() (*Memstore, error) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	return &Memstore{
		db: db,
	}, nil
}

// Create inserts the passed Account into the Memstore,
// returning an ErrAccountAlreadyExists error if an Account
// with the same ID already exists in the Memstore.
func (m *Memstore) Create(ctx context.Context, account accounts.Account) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	exists, err := txn.First("account", "id", account.ID)
	if err != nil {
		return err
	}
	if exists != nil {
		return accounts.ErrAccountAlreadyExists
	}
	err = txn.Insert("account", &account)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// Get retrieves the Account specified by the passed ID from
// the Memstore, returning an ErrAccountNotFound error if no
// Account matches the passed ID.
func (m *Memstore) Get(ctx context.Context, id string) (accounts.Account, error) {
	txn := m.db.Txn(false)
	account, err := txn.First("account", "id", id)
	if err != nil {
		return accounts.Account{}, err
	}
	if account == nil {
		return accounts.Account{}, accounts.ErrAccountNotFound
	}
	return *account.(*accounts.Account), nil
}

// Update applies the passed Change to the Account that matches
// the specified ID in the Memstore, if any Account matches the
// specified ID in the Memstore.
func (m *Memstore) Update(ctx context.Context, id string, change accounts.Change) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	account, err := txn.First("account", "id", id)
	if err != nil {
		return err
	}
	if account == nil {
		return nil
	}
	updated := accounts.Apply(change, *account.(*accounts.Account))
	err = txn.Insert("account", &updated)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// Delete removes the Account that matches the specified ID from
// the Memstore, if any Account matches the specified ID in the
// Memstore.
func (m *Memstore) Delete(ctx context.Context, id string) error {
	txn := m.db.Txn(true)
	defer txn.Abort()
	exists, err := txn.First("account", "id", id)
	if err != nil {
		return err
	}
	if exists == nil {
		return nil
	}
	err = txn.Delete("account", exists)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

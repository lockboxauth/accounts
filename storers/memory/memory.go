package memory

import (
	"context"

	memdb "github.com/hashicorp/go-memdb"

	"lockbox.dev/accounts"
)

var (
	schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"account": {
				Name: "account",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "ID", Lowercase: true},
					},
					"profileID": {
						Name:    "profileID",
						Indexer: &memdb.StringFieldIndex{Field: "ProfileID", Lowercase: true},
					},
				},
			},
		},
	}
)

// Storer is an in-memory implementation of the Storer
// interface.
type Storer struct {
	db *memdb.MemDB
}

// NewStorer returns an in-memory Storer instance that is ready
// to be used as a Storer.
func NewStorer() (*Storer, error) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	return &Storer{
		db: db,
	}, nil
}

// Create inserts the passed Account into the Storer,
// returning an ErrAccountAlreadyExists error if an Account
// with the same ID already exists in the Storer.
func (s *Storer) Create(ctx context.Context, account accounts.Account) error {
	txn := s.db.Txn(true)
	defer txn.Abort()
	exists, err := txn.First("account", "id", account.ID)
	if err != nil {
		return err
	}
	if exists != nil {
		return accounts.ErrAccountAlreadyExists
	}
	if account.IsRegistration {
		exists, err = txn.First("account", "profileID", account.ProfileID)
		if err != nil {
			return err
		}
		if exists != nil {
			return accounts.ErrProfileIDAlreadyExists
		}
	}
	err = txn.Insert("account", &account)
	if err != nil {
		return err
	}
	txn.Commit()
	return nil
}

// Get retrieves the Account specified by the passed ID from
// the Storer, returning an ErrAccountNotFound error if no
// Account matches the passed ID.
func (s *Storer) Get(ctx context.Context, id string) (accounts.Account, error) {
	txn := s.db.Txn(false)
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
// the specified ID in the Storer, if any Account matches the
// specified ID in the Storer.
func (s *Storer) Update(ctx context.Context, id string, change accounts.Change) error {
	txn := s.db.Txn(true)
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
// the Storer, if any Account matches the specified ID in the
// Storer.
func (s *Storer) Delete(ctx context.Context, id string) error {
	txn := s.db.Txn(true)
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

// ListByProfile returns all the Accounts associated with the passed profile ID,
// sorted with the most recently used Accounts coming first.
func (s *Storer) ListByProfile(ctx context.Context, profileID string) ([]accounts.Account, error) {
	txn := s.db.Txn(false)
	var accts []accounts.Account
	acctIter, err := txn.Get("account", "profileID", profileID)
	if err != nil {
		return nil, err
	}
	for {
		acct := acctIter.Next()
		if acct == nil {
			break
		}
		accts = append(accts, *acct.(*accounts.Account))
	}
	accounts.ByLastUsedDesc(accts)
	return accts, nil
}

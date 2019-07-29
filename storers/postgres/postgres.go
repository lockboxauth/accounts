package postgres

import (
	"context"
	"database/sql"

	"darlinggo.co/pan"
	"github.com/lib/pq"

	"lockbox.dev/accounts"
)

//go:generate go-bindata -pkg migrations -o migrations/generated.go sql/

const (
	TestConnStringEnvVar = "PG_TEST_DB"
)

// Storer provides a PostgreSQL-backed implementation of the Storer
// interface.
type Storer struct {
	db *sql.DB
}

// NewStorer returns a Storer instance that is backed by the specified
// *sql.DB. The returned Storer instance is ready to be used as a Storer.
func NewStorer(ctx context.Context, conn *sql.DB) *Storer {
	return &Storer{db: conn}
}

// Create inserts the passed Account into the PostgreSQL database, returning
// an ErrAccountAlreadyExists error if the Account's ID already exists in the
// database.
func (s *Storer) Create(ctx context.Context, account accounts.Account) error {
	query := createSQL(ctx, toPostgres(account))
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if e, ok := err.(*pq.Error); ok {
		if e.Constraint == "accounts_pkey" {
			err = accounts.ErrAccountAlreadyExists
		}
		if e.Constraint == "unique_registration" {
			err = accounts.ErrProfileIDAlreadyExists
		}
	}
	return err
}

// Get retrieves the Account specified by the passed ID from the PostgreSQL
// database. If no Account matches the passed ID, an ErrAccountNotFound error
// is returned.
func (s *Storer) Get(ctx context.Context, id string) (accounts.Account, error) {
	query := getSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return accounts.Account{}, err
	}
	rows, err := s.db.Query(queryStr, query.Args()...)
	if err != nil {
		return accounts.Account{}, err
	}
	var account Account
	for rows.Next() {
		err = pan.Unmarshal(rows, &account)
		if err != nil {
			return accounts.Account{}, err
		}
	}
	if err = rows.Err(); err != nil {
		return accounts.Account{}, err
	}
	if account.ID == "" {
		return accounts.Account{}, accounts.ErrAccountNotFound
	}
	return fromPostgres(account), nil
}

// Update applies the passed Change to the Account in the PostgreSQL database
// that matches the specified ID, if any Account matches the specified ID.
func (s *Storer) Update(ctx context.Context, id string, change accounts.Change) error {
	if change.IsEmpty() {
		return nil
	}
	query := updateSQL(ctx, id, change)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes the Account that matches the passed ID from the PostgreSQL
// database, if any Account matches the passed ID.
func (s *Storer) Delete(ctx context.Context, id string) error {
	query := deleteSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = s.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}

// ListByProfile returns all the Accounts associated with the passed profile ID,
// sorted with the most recently used Accounts coming first.
func (s *Storer) ListByProfile(ctx context.Context, profileID string) ([]accounts.Account, error) {
	query := listByProfileSQL(ctx, profileID)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	var accts []accounts.Account
	for rows.Next() {
		var account Account
		err = pan.Unmarshal(rows, &account)
		if err != nil {
			return accts, err
		}
		accts = append(accts, fromPostgres(account))
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	accounts.ByLastUsedDesc(accts)
	return accts, nil
}

package storers

import (
	"context"
	"database/sql"

	"darlinggo.co/pan"

	"github.com/lib/pq"

	"impractical.co/auth/accounts"
)

// Postgres provides a PostgreSQL-backed implementation of the Storer
// interface.
type Postgres struct {
	db *sql.DB
}

// NewPostgres returns a Postgres instance that is backed by the specified
// *sql.DB. The returned Postgres instance is ready to be used as a Storer.
func NewPostgres(ctx context.Context, conn *sql.DB) *Postgres {
	return &Postgres{db: conn}
}

// Create inserts the passed Account into the PostgreSQL database, returning
// an ErrAccountAlreadyExists error if the Account's ID already exists in the
// database.
func (p *Postgres) Create(ctx context.Context, account accounts.Account) error {
	query := createSQL(ctx, toPostgres(account))
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = p.db.Exec(queryStr, query.Args()...)
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
func (p *Postgres) Get(ctx context.Context, id string) (accounts.Account, error) {
	query := getSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return accounts.Account{}, err
	}
	rows, err := p.db.Query(queryStr, query.Args()...)
	if err != nil {
		return accounts.Account{}, err
	}
	var account postgresAccount
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
func (p *Postgres) Update(ctx context.Context, id string, change accounts.Change) error {
	if change.IsEmpty() {
		return nil
	}
	query := updateSQL(ctx, id, change)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = p.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}

// Delete removes the Account that matches the passed ID from the PostgreSQL
// database, if any Account matches the passed ID.
func (p *Postgres) Delete(ctx context.Context, id string) error {
	query := deleteSQL(ctx, id)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = p.db.Exec(queryStr, query.Args()...)
	if err != nil {
		return err
	}
	return nil
}

// ListByProfile returns all the Accounts associated with the passed profile ID,
// sorted with the most recently used Accounts coming first.
func (p *Postgres) ListByProfile(ctx context.Context, profileID string) ([]accounts.Account, error) {
	query := listByProfileSQL(ctx, profileID)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return nil, err
	}
	rows, err := p.db.Query(queryStr, query.Args()...)
	if err != nil {
		return nil, err
	}
	var accts []accounts.Account
	for rows.Next() {
		var account postgresAccount
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

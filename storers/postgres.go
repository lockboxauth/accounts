package storers

import (
	"context"
	"database/sql"

	"darlinggo.co/pan"

	"github.com/lib/pq"

	"code.impractical.co/accounts"
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
	query := createSQL(ctx, account)
	queryStr, err := query.PostgreSQLString()
	if err != nil {
		return err
	}
	_, err = p.db.Exec(queryStr, query.Args()...)
	if e, ok := err.(*pq.Error); ok {
		if e.Constraint == "accounts_pkey" {
			err = accounts.ErrAccountAlreadyExists
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
	var account accounts.Account
	for rows.Next() {
		err = pan.Unmarshal(rows, &account)
		if err != nil {
			return account, err
		}
	}
	if err = rows.Err(); err != nil {
		return account, err
	}
	if account.ID == "" {
		return account, accounts.ErrAccountNotFound
	}
	return account, nil
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

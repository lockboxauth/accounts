package storers

import (
	"context"
	"database/sql"

	"darlinggo.co/pan"

	"github.com/lib/pq"

	"code.impractical.co/accounts"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(ctx context.Context, conn *sql.DB) *Postgres {
	return &Postgres{db: conn}
}

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

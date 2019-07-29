package postgres

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"log"
	"net/url"
	"os"
	"sync"

	uuid "github.com/hashicorp/go-uuid"
	migrate "github.com/rubenv/sql-migrate"

	"lockbox.dev/accounts"
	"lockbox.dev/accounts/storers/postgres/migrations"
)

type Factory struct {
	db        *sql.DB
	databases map[string]*sql.DB
	lock      sync.Mutex
}

func NewFactory(db *sql.DB) *Factory {
	return &Factory{
		db:        db,
		databases: map[string]*sql.DB{},
	}
}

func (p *Factory) NewStorer(ctx context.Context) (accounts.Storer, error) {
	u, err := url.Parse(os.Getenv(TestConnStringEnvVar))
	if err != nil {
		log.Printf("Error parsing %s as a URL: %+v\n", TestConnStringEnvVar, err)
		return nil, err
	}
	if u.Scheme != "postgres" {
		return nil, errors.New(TestConnStringEnvVar + " must begin with postgres://")
	}

	tableSuffix, err := uuid.GenerateRandomBytes(6)
	if err != nil {
		log.Printf("Error generating table suffix: %+v\n", err)
		return nil, err
	}
	table := "accounts_test_" + hex.EncodeToString(tableSuffix)

	_, err = p.db.Exec("CREATE DATABASE " + table + ";")
	if err != nil {
		log.Printf("Error creating database %s: %+v\n", table, err)
		return nil, err
	}

	u.Path = "/" + table
	newConn, err := sql.Open("postgres", u.String())
	if err != nil {
		log.Println("Accidentally orphaned", table, "it will need to be cleaned up manually")
		return nil, err
	}

	p.lock.Lock()
	p.databases[table] = newConn
	p.lock.Unlock()

	migrations := &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "sql",
	}
	_, err = migrate.Exec(newConn, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, err
	}

	storer := NewStorer(ctx, newConn)

	return storer, nil
}

func (p *Factory) TeardownStorers() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	for table, conn := range p.databases {
		conn.Close()
		_, err := p.db.Exec("DROP DATABASE " + table + ";")
		if err != nil {
			return err
		}
	}
	p.db.Close()
	return nil
}

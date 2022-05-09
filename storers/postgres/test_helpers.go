package postgres

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"

	uuid "github.com/hashicorp/go-uuid"
	migrate "github.com/rubenv/sql-migrate"

	"lockbox.dev/accounts"
	"lockbox.dev/accounts/storers/postgres/migrations"
)

// Factory is a generator of Storers for testing purposes. It knows how to
// create, track, and clean up PostgreSQL databases that tests can be run
// against.
type Factory struct {
	db        *sql.DB
	databases map[string]*sql.DB
	lock      sync.Mutex
}

// NewFactory returns a Factory that is ready to be used. The passed sql.DB
// will be used as a control plane connection, but each test will have its own
// database created for that test.
func NewFactory(db *sql.DB) *Factory {
	return &Factory{
		db:        db,
		databases: map[string]*sql.DB{},
	}
}

// NewStorer retrieves the connection string from the environment (using
// TestConnStringEnvVar), parses it, and injects a new database name into it.
// The new database name is a random name prefixed with accounts_test_, and it
// will be automatically created in NewStorer. NewStorer also runs migrations,
// and keeps track of these test databases so they can be deleted automatically
// later.
func (p *Factory) NewStorer(ctx context.Context) (accounts.Storer, error) { //nolint:ireturn // interface requires returning an interface
	connString, err := url.Parse(os.Getenv(TestConnStringEnvVar))
	if err != nil {
		log.Printf("Error parsing %s as a URL: %+v\n", TestConnStringEnvVar, err)
		return nil, err
	}
	if connString.Scheme != "postgres" {
		return nil, fmt.Errorf("%s must begin with postgres://", TestConnStringEnvVar) //nolint:goerr113 // user facing error, doesn't get handled
	}

	tableSuffix, err := uuid.GenerateRandomBytes(6) //nolint:gomnd // number chosen arbitrarily, doesn't matter
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

	connString.Path = "/" + table
	newConn, err := sql.Open("postgres", connString.String())
	if err != nil {
		log.Println("Accidentally orphaned", table, "it will need to be cleaned up manually")
		return nil, err
	}

	p.lock.Lock()
	p.databases[table] = newConn
	p.lock.Unlock()

	migs := &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "sql",
	}
	_, err = migrate.Exec(newConn, "postgres", migs, migrate.Up)
	if err != nil {
		return nil, err
	}

	storer := NewStorer(ctx, newConn)

	return storer, nil
}

// TeardownStorers automatically deletes all the tracked databases created by
// NewStorer.
func (p *Factory) TeardownStorers() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	for table, conn := range p.databases {
		err := conn.Close()
		if err != nil {
			return err
		}
		_, err = p.db.Exec("DROP DATABASE " + table + ";")
		if err != nil {
			return err
		}
	}
	return p.db.Close()
}

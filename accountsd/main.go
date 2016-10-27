package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"code.impractical.co/accounts"
	"code.impractical.co/accounts/apiv1"
	"code.impractical.co/accounts/migrations"
	"code.impractical.co/accounts/storers"

	"github.com/rubenv/sql-migrate"

	"darlinggo.co/healthcheck"
	"darlinggo.co/version"
)

func main() {
	// Set up our logger
	logger := log.New(os.Stdout, "", log.Llongfile|log.LstdFlags|log.LUTC|log.Lmicroseconds)

	// Set up postgres connection
	postgres := os.Getenv("PG_DB")
	if postgres == "" {
		logger.Println("Error setting up Postgres: no connection string set.")
		os.Exit(1)
	}
	db, err := sql.Open("postgres", postgres)
	if err != nil {
		logger.Printf("Error connecting to Postgres: %+v\n", err)
		os.Exit(1)
	}

	// set up our base context
	ctx := context.Background()

	// run our postgres migrations
	migrations := &migrate.AssetMigrationSource{
		Asset:    migrations.Asset,
		AssetDir: migrations.AssetDir,
		Dir:      "sql",
	}
	_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		logger.Printf("Error running migrations for Postgres: %+v\n", err)
		os.Exit(1)
	}

	// build our APIv1 struct
	storer := storers.NewPostgres(ctx, db)
	v1 := apiv1.APIv1{
		Dependencies: accounts.Dependencies{
			Storer: storer,
			Log:    logger,
		},
	}

	// set up our APIv1 handlers
	// we need both to avoid redirecting, which turns POST into GET
	// the slash is needed to handle /v1/
	http.Handle("/v1/", v1.Server("/v1/"))
	http.Handle("/v1", v1.Server("/v1"))

	// set up version handler
	http.Handle("/version", version.Handler)

	// set up health check
	dbCheck := healthcheck.NewSQL(db, "main Postgres DB")
	checker := healthcheck.NewChecks(ctx, logger.Printf, dbCheck)
	http.Handle("/health", checker)

	// make our version information pretty
	vers := version.Tag
	if vers == "undefined" || vers == "" {
		vers = "dev"
	}
	vers = vers + " (" + version.Hash + ")"

	// let users know what's listening and on what address/port
	logger.Printf("accountsd version %s starting on port 0.0.0.0:4003\n", vers)
	err = http.ListenAndServe("0.0.0.0:4003", nil)
	if err != nil {
		logger.Printf("Error listening on port 0.0.0.0:4003: %+v\n", err)
		os.Exit(1)
	}
}

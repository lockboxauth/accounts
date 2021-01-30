// Package postgres provides an implementation of the
// lockbox.dev/accounts.Storer interface that stores data in a PostgreSQL
// database.
//
// The package assumes that the database is set up and ready for its use, and
// does not automatically set up the database itself. Migrations to set the
// database up are available in the sql folder of this package. The migrations
// are also available in the migrations package, which contains the contents of
// the sql folder packaged using go-bindata to make them easy to include in Go
// binaries. Migrations should be applied in lexicographical order, with
// numbers coming before letters.
package postgres

// Package memory provides an in-memory implementation of the
// lockbox.dev/accounts.Storer interface.
//
// This implementation is useful for testing and demo setups in which data is
// not meant to be stored reliably or for a long time. All the data will be
// permanently lost when the service process exits.
package memory

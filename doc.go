// Package accounts provides a mapping of login methods to logical users.
//
// The accounts package provides the definitions of the service and its
// boundaries. It sets up the Account type, which represents a mapping of a
// login method to a user within your application, and the Storer interface,
// which defines how to implement data storage backends for these Accounts.
//
// This package can be thought of as providing the types and helpers that form
// the conceptual framework of the subsystem, but with very little
// functionality provided by itself. Instead, implementations of the interfaces
// and sub-packages using these types are where most functionality will
// actually live.
package accounts

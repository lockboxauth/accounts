// Package apiv1 provides a JSON API for interacting with accounts.
//
// This package can be imported to get an http.Handler that will provide access
// for retrieving Accounts, listing Accounts by their ProfileID, adding
// Accounts, and deleting Accounts.
//
// The lockbox.dev/sessions package is used to authenticate a JWT bearer token
// for deleting Accounts, retrieving a specific Account, adding new Accounts to
// an existing profile, or listing Accounts associated with a profile. The
// bearer token's AccountID will be used as an Account's ID, and that Account's
// ProfileID must match the ProfileID of the Accounts being acted on or the
// profile Accounts are being listed for.
package apiv1

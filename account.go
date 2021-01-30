package accounts

import (
	"errors"
	"sort"
	"time"
)

var (
	// ErrAccountNotFound is returned when an Account was expected but could not be found.
	ErrAccountNotFound = errors.New("account not found")
	// ErrAccountAlreadyExists is returned when attempting to create an Account that already exists.
	ErrAccountAlreadyExists = errors.New("account already exists")
	// ErrProfileIDAlreadyExists is returned when an account is registered by the ProfileID already exists.
	ErrProfileIDAlreadyExists = errors.New("profileID already exists")
)

// Account is a representation of a user's identifier. It maps
// the identifier (email, username, whatever) to a profile ID,
// allowing users to have multiple identifiers that are all
// interchangeable.
type Account struct {
	// ID is a globally-unique identifier for how the user identifies
	// themselves to your application. It is case-insensitive.
	ID string

	// ProfileID is how your application should identify the user. It is an
	// opaque string that will be automatically generated for you.
	ProfileID string

	// Created is the time at which the Account was first registered.
	Created time.Time

	// LastUsed is the time at which the Account was last authenticated
	// with by the user, completing the password challenge or clicking the
	// link in the email, or however the account is authenticated.
	LastUsed time.Time

	// LastSeen is the time at which the Account was last seen acting. This
	// is different from the time it was last authenticated; when an
	// authentication token issued for this Account is used, LastSeen
	// should be updated. When the Account is issued an authentication
	// token, LastUsed and LastSeen should both be updated.
	LastSeen time.Time

	// IsRegistration should be set to true when the Account is the first
	// Account a user is trying to register. This enables extra validation
	// logic to ensure that ProfileIDs are unique for logical users, but
	// that multiple Accounts can be registered to a single logical user.
	IsRegistration bool
}

// Change represents a requested change to one or more of an
// Account's mutable properties.
type Change struct {
	LastUsed *time.Time
	LastSeen *time.Time
}

// IsEmpty returns true if the Change would not result in a
// change, no matter which Account it was applied to.
func (c Change) IsEmpty() bool {
	if c.LastUsed != nil {
		return false
	}
	if c.LastSeen != nil {
		return false
	}
	return true
}

// Apply returns a copy of the specified Account with the
// changes requested by the specified Change applied.
func Apply(change Change, account Account) Account {
	if change.IsEmpty() {
		return account
	}
	res := account
	if change.LastUsed != nil {
		res.LastUsed = *change.LastUsed
	}
	if change.LastSeen != nil {
		res.LastSeen = *change.LastSeen
	}
	return res
}

// FillDefaults sets a reasonable default for any of the properties
// of the specified Account that both have reasonable defaults and
// are set to the zero value when FillDefaults is called. It returns
// a copy of the specified Account with those defaults applied.
func FillDefaults(account Account) Account {
	res := account
	if res.Created.IsZero() {
		res.Created = time.Now()
	}
	if res.LastUsed.IsZero() {
		res.LastUsed = res.Created
	}
	if res.LastSeen.IsZero() {
		res.LastSeen = res.LastUsed
	}
	return res
}

// Dependencies holds all the information that we want to make available
// to all our functions, but that are orthogonal enough to not warrant
// their own place in every function's signature.
type Dependencies struct {
	Storer Storer
}

// ByLastUsedDesc sorts the passed slice of Accounts by their LastUsed
// property, with the most recent times at the lower indices.
func ByLastUsedDesc(accounts []Account) {
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].LastUsed.After(accounts[j].LastUsed) })
}

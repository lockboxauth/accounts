package accounts

//go:generate go-bindata -pkg migrations -o migrations/generated.go sql/

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/apex/log"
)

var (
	// ErrAccountNotFound is returned when an Account was expected but could not be found.
	ErrAccountNotFound = errors.New("account not found")
	// ErrAccountAlreadyExists is returned when attempting to create an Account that already exists.
	ErrAccountAlreadyExists = errors.New("account already exists")
)

// Account is a representation of a user's identifier. It maps
// the identifier (email, username, whatever) to a profile ID,
// allowing users to have multiple identifiers that are all
// interchangeable.
type Account struct {
	ID        string
	ProfileID string
	Created   time.Time
	LastUsed  time.Time
	LastSeen  time.Time
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

// Storer dictates how Accounts will be persisted and how to
// interact with those persisted Accounts.
type Storer interface {
	Create(ctx context.Context, account Account) error
	Get(ctx context.Context, id string) (Account, error)
	Update(ctx context.Context, id string, change Change) error
	Delete(ctx context.Context, id string) error
	ListByProfile(ctx context.Context, profileID string) ([]Account, error)
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
	Log    *log.Logger
}

type byLastUsedDesc []Account

func (b byLastUsedDesc) Less(i, j int) bool { return b[i].LastUsed.After(b[j].LastUsed) }

func (b byLastUsedDesc) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b byLastUsedDesc) Len() int { return len(b) }

// ByLastUsedDesc sorts the passed slice of Accounts by their LastUsed
// property, with the most recent times at the lower indices.
func ByLastUsedDesc(accounts []Account) {
	sort.Sort(byLastUsedDesc(accounts))
}

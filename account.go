package accounts

//go:generate go-bindata -pkg migrations -o migrations/generated.go sql/

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
)

type Account struct {
	ID        string    `sql_column:"id"`
	ProfileID string    `sql_column:"profile_id"`
	Created   time.Time `sql_column:"created_at"`
	LastUsed  time.Time `sql_column:"last_used_at"`
	LastSeen  time.Time `sql_column:"last_seen_at"`
}

func (a Account) GetSQLTableName() string {
	return "accounts"
}

type Change struct {
	LastUsed *time.Time
	LastSeen *time.Time
}

func (c Change) IsEmpty() bool {
	if c.LastUsed != nil {
		return false
	}
	if c.LastSeen != nil {
		return false
	}
	return true
}

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

type Storer interface {
	Create(ctx context.Context, account Account) error
	Get(ctx context.Context, id string) (Account, error)
	Update(ctx context.Context, id string, change Change) error
	Delete(ctx context.Context, id string) error
}

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

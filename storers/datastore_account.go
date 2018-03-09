package storers

import (
	"time"

	"impractical.co/auth/accounts"
)

type datastoreAccount struct {
	ID        string `datastore:"-"`
	ProfileID string
	Created   time.Time
	LastUsed  time.Time
	LastSeen  time.Time
}

func fromDatastore(a datastoreAccount) accounts.Account {
	return accounts.Account(a)
}

func toDatastore(a accounts.Account) datastoreAccount {
	return datastoreAccount(a)
}

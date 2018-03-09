package storers

import (
	"time"

	"impractical.co/auth/accounts"
)

type postgresAccount struct {
	ID        string    `sql_column:"id"`
	ProfileID string    `sql_column:"profile_id"`
	Created   time.Time `sql_column:"created_at"`
	LastUsed  time.Time `sql_column:"last_used_at"`
	LastSeen  time.Time `sql_column:"last_seen_at"`
}

func fromPostgres(a postgresAccount) accounts.Account {
	return accounts.Account(a)
}

func toPostgres(a accounts.Account) postgresAccount {
	return postgresAccount(a)
}

func (p postgresAccount) GetSQLTableName() string {
	return "accounts"
}

package postgres

import (
	"database/sql"
	"time"

	"lockbox.dev/accounts"
)

// Account is a representation of the accounts.Account type that is suitable to
// be stored in a PostgreSQL database.
type Account struct {
	ID             string       `sql_column:"id"`
	ProfileID      string       `sql_column:"profile_id"`
	Created        time.Time    `sql_column:"created_at"`
	LastUsed       time.Time    `sql_column:"last_used_at"`
	LastSeen       time.Time    `sql_column:"last_seen_at"`
	IsRegistration sql.NullBool `sql_column:"is_registration"`
}

func fromPostgres(account Account) accounts.Account {
	acct := accounts.Account{
		ID:        account.ID,
		ProfileID: account.ProfileID,
		Created:   account.Created,
		LastUsed:  account.LastUsed,
		LastSeen:  account.LastSeen,
	}
	if account.IsRegistration.Valid {
		acct.IsRegistration = account.IsRegistration.Bool
	}
	return acct
}

func toPostgres(account accounts.Account) Account {
	return Account{
		ID:        account.ID,
		ProfileID: account.ProfileID,
		Created:   account.Created,
		LastUsed:  account.LastUsed,
		LastSeen:  account.LastSeen,
		IsRegistration: sql.NullBool{
			Valid: account.IsRegistration,
			Bool:  account.IsRegistration,
		},
	}
}

// GetSQLTableName returns the name of the SQL table that the data for this
// type will be stored in.
func (Account) GetSQLTableName() string {
	return "accounts"
}

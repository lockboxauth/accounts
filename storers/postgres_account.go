package storers

import (
	"database/sql"
	"time"

	"impractical.co/auth/accounts"
)

type postgresAccount struct {
	ID             string       `sql_column:"id"`
	ProfileID      string       `sql_column:"profile_id"`
	Created        time.Time    `sql_column:"created_at"`
	LastUsed       time.Time    `sql_column:"last_used_at"`
	LastSeen       time.Time    `sql_column:"last_seen_at"`
	IsRegistration sql.NullBool `sql_column:"is_registration"`
}

func fromPostgres(a postgresAccount) accounts.Account {
	acct := accounts.Account{
		ID:        a.ID,
		ProfileID: a.ProfileID,
		Created:   a.Created,
		LastUsed:  a.LastUsed,
		LastSeen:  a.LastSeen,
	}
	if a.IsRegistration.Valid {
		acct.IsRegistration = a.IsRegistration.Bool
	}
	return acct
}

func toPostgres(a accounts.Account) postgresAccount {
	return postgresAccount{
		ID:        a.ID,
		ProfileID: a.ProfileID,
		Created:   a.Created,
		LastUsed:  a.LastUsed,
		LastSeen:  a.LastSeen,
		IsRegistration: sql.NullBool{
			Valid: a.IsRegistration,
			Bool:  a.IsRegistration,
		},
	}
}

func (p postgresAccount) GetSQLTableName() string {
	return "accounts"
}

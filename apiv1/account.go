package apiv1

import (
	"time"

	"lockbox.dev/accounts"
)

// Account is the API representation of an Account.
// it dictates what the JSON representation of Accounts
// will be.
type Account struct {
	ID             string    `json:"id"`
	ProfileID      string    `json:"profileID"`
	IsRegistration bool      `json:"isRegistration"`
	CreatedAt      time.Time `json:"createdAt"`
	LastSeenAt     time.Time `json:"lastSeenAt,omitempty"`
	LastUsedAt     time.Time `json:"lastUsedAt,omitempty"`
}

// Change is the API representation of a Change.
// It dictates what the JSON representation of Changes
// will be.
type Change struct {
	LastSeenAt *time.Time `json:"lastSeenAt,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
}

func coreAccount(account Account) accounts.Account {
	return accounts.Account{
		ID:             account.ID,
		ProfileID:      account.ProfileID,
		IsRegistration: account.IsRegistration,
		Created:        account.CreatedAt,
		LastSeen:       account.LastSeenAt,
		LastUsed:       account.LastUsedAt,
	}
}

func coreAccounts(accts []Account) []accounts.Account {
	res := make([]accounts.Account, 0, len(accts))
	for _, acct := range accts {
		res = append(res, coreAccount(acct))
	}
	return res
}

func apiAccount(account accounts.Account) Account {
	return Account{
		ID:             account.ID,
		ProfileID:      account.ProfileID,
		IsRegistration: account.IsRegistration,
		CreatedAt:      account.Created,
		LastSeenAt:     account.LastSeen,
		LastUsedAt:     account.LastUsed,
	}
}

func apiAccounts(accts []accounts.Account) []Account {
	res := make([]Account, 0, len(accts))
	for _, acct := range accts {
		res = append(res, apiAccount(acct))
	}
	return res
}

func coreChange(change Change) accounts.Change {
	return accounts.Change{
		LastSeen: change.LastSeenAt,
		LastUsed: change.LastUsedAt,
	}
}

func apiChange(change accounts.Change) Change {
	return Change{
		LastSeenAt: change.LastSeen,
		LastUsedAt: change.LastUsed,
	}
}

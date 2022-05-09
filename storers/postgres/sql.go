package postgres

import (
	"context"

	"darlinggo.co/pan"

	"lockbox.dev/accounts"
)

func getSQL(_ context.Context, id string) *pan.Query {
	var account Account
	q := pan.New("SELECT " + pan.Columns(account).String() + " FROM " + pan.Table(account))
	q.Where()
	q.Comparison(account, "ID", "=", id)
	return q.Flush(" ")
}

func createSQL(_ context.Context, account Account) *pan.Query {
	return pan.Insert(account)
}

func updateSQL(_ context.Context, id string, change accounts.Change) *pan.Query {
	var account Account
	query := pan.New("UPDATE " + pan.Table(account) + " SET ")
	if change.LastUsed != nil {
		query.Comparison(account, "LastUsed", "=", *change.LastUsed)
	}
	if change.LastSeen != nil {
		query.Comparison(account, "LastSeen", "=", *change.LastSeen)
	}
	query.Flush(", ")
	query.Where()
	query.Comparison(account, "ID", "=", id)
	return query.Flush(" ")
}

func deleteSQL(_ context.Context, id string) *pan.Query {
	var account Account
	q := pan.New("DELETE FROM " + pan.Table(account))
	q.Where()
	q.Comparison(account, "ID", "=", id)
	return q.Flush(" ")
}

func listByProfileSQL(_ context.Context, profileID string) *pan.Query {
	var account Account
	q := pan.New("SELECT " + pan.Columns(account).String() + " FROM " + pan.Table(account))
	q.Where()
	q.Comparison(account, "ProfileID", "=", profileID)
	q.OrderByDesc("last_used_at")
	return q.Flush(" ")
}

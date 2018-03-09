package storers

import (
	"context"

	"darlinggo.co/pan"
	"impractical.co/auth/accounts"
)

func getSQL(ctx context.Context, id string) *pan.Query {
	var account postgresAccount
	q := pan.New("SELECT " + pan.Columns(account).String() + " FROM " + pan.Table(account))
	q.Where()
	q.Comparison(account, "ID", "=", id)
	return q.Flush(" ")
}

func createSQL(ctx context.Context, account postgresAccount) *pan.Query {
	return pan.Insert(account)
}

func updateSQL(ctx context.Context, id string, change accounts.Change) *pan.Query {
	var account postgresAccount
	q := pan.New("UPDATE " + pan.Table(account) + " SET ")
	if change.LastUsed != nil {
		q.Comparison(account, "LastUsed", "=", *change.LastUsed)
	}
	if change.LastSeen != nil {
		q.Comparison(account, "LastSeen", "=", *change.LastSeen)
	}
	q.Flush(", ")
	q.Where()
	q.Comparison(account, "ID", "=", id)
	return q.Flush(" ")
}

func deleteSQL(ctx context.Context, id string) *pan.Query {
	var account postgresAccount
	q := pan.New("DELETE FROM " + pan.Table(account))
	q.Where()
	q.Comparison(account, "ID", "=", id)
	return q.Flush(" ")
}

func listByProfileSQL(ctx context.Context, profileID string) *pan.Query {
	var account postgresAccount
	q := pan.New("SELECT " + pan.Columns(account).String() + " FROM " + pan.Table(account))
	q.Where()
	q.Comparison(account, "ProfileID", "=", profileID)
	q.OrderByDesc("last_used_at")
	return q.Flush(" ")
}

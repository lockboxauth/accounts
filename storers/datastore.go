package storers

import (
	"context"

	"cloud.google.com/go/datastore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	yall "yall.in"

	"impractical.co/auth/accounts"
)

const datastoreAccountKind = "Account"

// Datastore provides a Google Cloud Datastore-backed implementation of the
// Storer interface.
type Datastore struct {
	client    *datastore.Client
	namespace string
}

// NewDatastore returns a Datastore instance that is backed by the specified
// *datastore.Client. The returned Datastore instance is ready to be used as a
// Storer.
func NewDatastore(ctx context.Context, client *datastore.Client) *Datastore {
	return &Datastore{client: client}
}

func (d *Datastore) key(id string) *datastore.Key {
	key := datastore.NameKey(datastoreAccountKind, id, nil)
	if d.namespace != "" {
		key.Namespace = d.namespace
	}
	return key
}

// Create inserts the passed Account into the Datastore, returning an
// ErrAccountAlreadyExists error if the Account's ID already exists in the
// database.
func (d *Datastore) Create(ctx context.Context, account accounts.Account) error {
	acct := toDatastore(account)
	mut := datastore.NewInsert(d.key(acct.ID), &acct)
	_, err := d.client.Mutate(ctx, mut)
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return accounts.ErrAccountAlreadyExists
		}
		return err
	}
	return nil
}

// Get retrieves the Account specified by the passed ID from the PostgreSQL
// database. If no Account matches the passed ID, an ErrAccountNotFound error
// is returned.
func (d *Datastore) Get(ctx context.Context, id string) (accounts.Account, error) {
	var account datastoreAccount
	err := d.client.Get(ctx, d.key(id), &account)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return accounts.Account{}, accounts.ErrAccountNotFound
		}
		return accounts.Account{}, err
	}
	account.ID = id
	return fromDatastore(account), nil
}

// Update applies the passed Change to the Account in the PostgreSQL database
// that matches the specified ID, if any Account matches the specified ID.
func (d *Datastore) Update(ctx context.Context, id string, change accounts.Change) error {
	if change.IsEmpty() {
		return nil
	}
	_, err := d.client.RunInTransaction(ctx, func(t *datastore.Transaction) error {
		var account datastoreAccount
		err := t.Get(d.key(id), &account)
		if err == datastore.ErrNoSuchEntity {
			return nil
		} else if err != nil {
			return err
		}
		acct := toDatastore(accounts.Apply(change, fromDatastore(account)))
		_, err = t.Put(d.key(id), &acct)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Delete removes the Account that matches the passed ID from the PostgreSQL
// database, if any Account matches the passed ID.
func (d *Datastore) Delete(ctx context.Context, id string) error {
	return d.client.Delete(ctx, d.key(id))
}

// ListByProfile returns all the Accounts associated with the passed profile ID,
// sorted with the most recently used Accounts coming first.
func (d *Datastore) ListByProfile(ctx context.Context, profileID string) ([]accounts.Account, error) {
	q := datastore.NewQuery(datastoreAccountKind).Filter("ProfileID =", profileID).KeysOnly()
	if d.namespace != "" {
		q = q.Namespace(d.namespace)
	}
	keys, err := d.client.GetAll(ctx, q, nil)
	if err != nil {
		return nil, err
	}
	yall.FromContext(ctx).WithField("num_keys", len(keys)).WithField("namespace", d.namespace).Debug("got keys")
	for _, key := range keys {
		yall.FromContext(ctx).WithField("id", key.Name).Debug("got key")
	}
	if len(keys) == 0 {
		return nil, nil
	}
	accts := make([]datastoreAccount, len(keys))
	err = d.client.GetMulti(ctx, keys, accts)
	if err != nil {
		return nil, err
	}
	results := make([]accounts.Account, 0, len(accts))
	for pos, acct := range accts {
		acct.ID = keys[pos].Name
		results = append(results, fromDatastore(acct))
	}
	accounts.ByLastUsedDesc(results)
	return results, nil
}

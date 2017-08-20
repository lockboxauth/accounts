package storers

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"impractical.co/auth/accounts"

	"github.com/pborman/uuid"
)

const (
	changeLastUsed = 1 << iota
	changeLastSeen
	changeVariations
)

type StorerFactory interface {
	NewStorer(ctx context.Context) (accounts.Storer, error)
	TeardownStorers() error
}

var storerFactories []StorerFactory

func compareAccounts(account1, account2 accounts.Account) (success bool, field string, val1, val2 interface{}) {
	if account1.ID != account2.ID {
		return false, "ID", account1.ID, account2.ID
	}
	if account1.ProfileID != account2.ProfileID {
		return false, "ProfileID", account1.ProfileID, account2.ProfileID
	}
	if !account1.Created.Equal(account2.Created) {
		return false, "Created", account1.Created, account2.Created
	}
	if !account1.LastUsed.Equal(account2.LastUsed) {
		return false, "LastUsed", account1.LastUsed, account2.LastUsed
	}
	if !account1.LastSeen.Equal(account2.LastSeen) {
		return false, "LastSeen", account1.LastSeen, account2.LastSeen
	}
	return true, "", nil, nil
}

func TestMain(m *testing.M) {
	flag.Parse()
	result := m.Run()
	for _, factory := range storerFactories {
		err := factory.TeardownStorers()
		if err != nil {
			log.Printf("Error cleaning up after %T: %+v\n", factory, err)
		}
	}
	os.Exit(result)
}

func runTest(t *testing.T, f func(*testing.T, accounts.Storer, context.Context)) {
	t.Parallel()
	for _, factory := range storerFactories {
		ctx := context.Background()
		storer, err := factory.NewStorer(ctx)
		if err != nil {
			t.Fatalf("Error creating Storer from %T: %+v\n", factory, err)
		}
		t.Run(fmt.Sprintf("Storer=%T", storer), func(t *testing.T) {
			t.Parallel()
			f(t, storer, ctx)
		})
	}
}

func TestCreateAndGetAccount(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuid.NewRandom().String(),
			Created:   time.Now().Round(time.Millisecond),
			LastUsed:  time.Now().Round(time.Millisecond),
			LastSeen:  time.Now().Round(time.Millisecond),
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}

		resp, err := storer.Get(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}
		ok, field, expected, result := compareAccounts(account, resp)
		if !ok {
			t.Errorf("Expected %s to be %v, got %v\n", field, expected, result)
		}
	})
}

func TestGetNonexistentAccount(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		_, err := storer.Get(ctx, "myaccount@impractical.co")
		if err != accounts.ErrAccountNotFound {
			t.Fatalf("Expected ErrAccountNotFound, got %v\n", err)
		}
	})
}

func TestCreateDuplicateID(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuid.NewRandom().String(),
			Created:   time.Now().Round(time.Millisecond),
			LastUsed:  time.Now().Round(time.Millisecond),
			LastSeen:  time.Now().Round(time.Millisecond),
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}
		account2 := accounts.Account{
			ID:        account.ID,
			ProfileID: uuid.NewRandom().String(),
			Created:   time.Now().Add(time.Hour).Round(time.Millisecond),
			LastUsed:  time.Now().Add(time.Hour).Round(time.Millisecond),
			LastSeen:  time.Now().Add(time.Hour).Round(time.Millisecond),
		}

		err = storer.Create(ctx, account2)
		if err != accounts.ErrAccountAlreadyExists {
			t.Fatalf("Expected ErrAccountAlreadyExists, got %v\n", err)
		}

		// we shouldn't have changed anything about what was stored
		resp, err := storer.Get(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}

		ok, field, expected, result := compareAccounts(account, resp)
		if !ok {
			t.Errorf("Expected %s to be %v, got %v\n", field, expected, result)
		}
	})
}

func TestCreateMultipleAccounts(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuid.NewRandom().String(),
			Created:   time.Now().Round(time.Millisecond),
			LastUsed:  time.Now().Round(time.Millisecond),
			LastSeen:  time.Now().Round(time.Millisecond),
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}
		account2 := account
		account2.ID = "paddy@impracticallabs.com"
		err = storer.Create(ctx, account2)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}

		resp, err := storer.Get(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}
		ok, field, expected, result := compareAccounts(account, resp)
		if !ok {
			t.Errorf("Expected %s to be %v, got %v\n", field, expected, result)
		}

		resp, err = storer.Get(ctx, account2.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}
		ok, field, expected, result = compareAccounts(account2, resp)
		if !ok {
			t.Errorf("Expected %s to be %v, got %v\n", field, expected, result)
		}
	})
}

func TestListAccountsByProfile(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuid.NewRandom().String(),
			Created:   time.Now().Round(time.Millisecond),
			LastUsed:  time.Now().Round(time.Millisecond),
			LastSeen:  time.Now().Round(time.Millisecond),
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}
		account2 := account
		account2.ID = "paddy@impracticallabs.com"
		account2.LastUsed = account2.LastUsed.Add(-1 * time.Minute)
		err = storer.Create(ctx, account2)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}

		accounts, err := storer.ListByProfile(ctx, account.ProfileID)
		if err != nil {
			t.Fatalf("Unexpected error listing accounts: %+v\n", err)
		}

		if len(accounts) != 2 {
			t.Fatalf("Expected %d accounts, got %d: %+v\n", 2, len(accounts), accounts)
		}
		ok, field, exp, res := compareAccounts(accounts[0], account)
		if !ok {
			t.Errorf("Expected %s to be %v for %s, got %v\n", field, exp, account.ID, res)
		}
		ok, field, exp, res = compareAccounts(accounts[1], account2)
		if !ok {
			t.Errorf("Expected %s to be %v for %s, got %v\n", field, exp, account2.ID, res)
		}
	})
}

func TestUpdateOneOfMany(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		for i := 1; i < changeVariations; i++ {
			i := i
			t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
				t.Parallel()

				account := accounts.Account{
					ID:        fmt.Sprintf("paddy+%d@impractical.co", i),
					ProfileID: uuid.NewRandom().String(),
					Created:   time.Now().Round(time.Millisecond),
					LastUsed:  time.Now().Round(time.Millisecond),
					LastSeen:  time.Now().Round(time.Millisecond),
				}
				err := storer.Create(ctx, account)
				if err != nil {
					t.Fatalf("Unexpected error creating account: %+v\n", err)

				}

				var throwaways []accounts.Account
				for x := 0; x < 5; x++ {
					throwaway := accounts.Account{
						ID:        fmt.Sprintf("paddy+%d+%d@impractical.co", i, x),
						ProfileID: uuid.NewRandom().String(),
						Created:   time.Now().Add(time.Duration(x) * time.Minute).Round(time.Millisecond),
						LastUsed:  time.Now().Add(time.Duration(x) * time.Hour).Round(time.Millisecond),
						LastSeen:  time.Now().Add(time.Duration(x) * time.Second).Round(time.Millisecond),
					}
					if x%2 == 0 {
						throwaway.ProfileID = account.ProfileID
					}

					err := storer.Create(ctx, throwaway)
					if err != nil {
						t.Fatalf("Unexpected error creating account: %+v\n", err)
					}
					throwaways = append(throwaways, throwaway)
				}

				var change accounts.Change
				if i&changeLastSeen != 0 {
					seen := time.Now().Add(time.Duration(i) * time.Minute).Round(time.Millisecond)
					change.LastSeen = &seen
				}
				if i&changeLastUsed != 0 {
					used := time.Now().Add(time.Duration(i) * time.Hour).Round(time.Millisecond)
					change.LastUsed = &used
				}
				expectation := accounts.Apply(change, account)

				err = storer.Update(ctx, account.ID, change)
				if err != nil {
					t.Fatalf("Unexpected error updating account: %+v\n", err)
				}
				result, err := storer.Get(ctx, account.ID)
				if err != nil {
					t.Fatalf("Unexpected error retrieving account: %+v\n", err)
				}
				ok, field, exp, res := compareAccounts(expectation, result)
				if !ok {
					t.Errorf("Expected %s to be %v, got %v\n", field, exp, res)
				}
				for _, throwaway := range throwaways {
					result, err := storer.Get(ctx, throwaway.ID)
					if err != nil {
						t.Errorf("Unexpected error retrieving account: %+v\n", err)
					}
					ok, field, exp, res := compareAccounts(throwaway, result)
					if !ok {
						t.Errorf("Expected %s to be %v for %s, got %v\n", field, exp, throwaway.ID, res)
					}
				}
			})
		}
	})
}

func TestUpdateNonExistent(t *testing.T) {
	// updating an account that doesn't exist is not an error
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		used := time.Now().Round(time.Millisecond)
		change := accounts.Change{
			LastUsed: &used,
		}
		err := storer.Update(ctx, "notanactualaccount@impractical.co", change)
		if err != nil {
			t.Fatalf("Unexpected error updating account: %+v\n", err)
		}
	})
}

func TestUpdateNoChange(t *testing.T) {
	// updating an account with an empty change should not error
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		var change accounts.Change
		err := storer.Update(ctx, "notanactualaccount@impractical.co", change)
		if err != nil {
			t.Fatalf("Unexpected error updating account: %+v\n", err)
		}
	})
}

func TestDeleteOneOfMany(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuid.NewRandom().String(),
			Created:   time.Now().Round(time.Millisecond),
			LastUsed:  time.Now().Round(time.Millisecond),
			LastSeen:  time.Now().Round(time.Millisecond),
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}

		var throwaways []accounts.Account
		for x := 0; x < 5; x++ {
			throwaway := accounts.Account{
				ID:        fmt.Sprintf("paddy+%d@impractical.co", x),
				ProfileID: uuid.NewRandom().String(),
				Created:   time.Now().Add(time.Duration(x) * time.Minute).Round(time.Millisecond),
				LastUsed:  time.Now().Add(time.Duration(x) * time.Hour).Round(time.Millisecond),
				LastSeen:  time.Now().Add(time.Duration(x) * time.Second).Round(time.Millisecond),
			}
			if x%2 == 0 {
				throwaway.ProfileID = account.ProfileID
			}

			err := storer.Create(ctx, throwaway)
			if err != nil {
				t.Fatalf("Unexpected error creating account: %+v\n", err)
			}
			throwaways = append(throwaways, throwaway)
		}

		err = storer.Delete(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error deleting account: %+v\n", err)
		}
		res, err := storer.Get(ctx, account.ID)
		if err != accounts.ErrAccountNotFound {
			t.Logf("Account: %+v\n", res)
			t.Errorf("Expected error to be ErrAccountNotFound, got %v\n", err)
		}
		for _, throwaway := range throwaways {
			result, err := storer.Get(ctx, throwaway.ID)
			if err != nil {
				t.Errorf("Unexpected error retrieving account: %+v\n", err)
			}
			ok, field, exp, res := compareAccounts(throwaway, result)
			if !ok {
				t.Errorf("Expected %s to be %v for %s, got %v\n", field, exp, throwaway.ID, res)
			}
		}
	})
}

func TestDeleteNonExistent(t *testing.T) {
	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		// we shouldn't get an error deleting an account that doesn't exist
		err := storer.Delete(ctx, "notarealaccount@impractical.co")
		if err != nil {
			t.Fatalf("Unexpected error deleting account: %+v\n", err)
		}
	})
}

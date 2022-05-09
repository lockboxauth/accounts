package accounts_test

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/hashicorp/go-uuid"
	yall "yall.in"
	"yall.in/colour"

	"lockbox.dev/accounts"
	"lockbox.dev/accounts/storers/memory"
	"lockbox.dev/accounts/storers/postgres"
)

const (
	changeLastUsed = 1 << iota
	changeLastSeen
	changeVariations
)

type Factory interface {
	NewStorer(ctx context.Context) (accounts.Storer, error)
	TeardownStorers() error
}

var factories []Factory

func uuidOrFail(t *testing.T) string {
	t.Helper()
	id, err := uuid.GenerateUUID()
	if err != nil {
		t.Fatalf("Unexpected error generating ID: %s", err.Error())
	}
	return id
}

func TestMain(m *testing.M) {
	flag.Parse()

	// set up our test storers
	factories = append(factories, memory.Factory{})
	if os.Getenv(postgres.TestConnStringEnvVar) != "" {
		storerConn, err := sql.Open("postgres", os.Getenv(postgres.TestConnStringEnvVar))
		if err != nil {
			panic(err)
		}
		factories = append(factories, postgres.NewFactory(storerConn))
	}

	// run the tests
	result := m.Run()

	// tear down all the storers we created
	for _, factory := range factories {
		err := factory.TeardownStorers()
		if err != nil {
			log.Printf("Error cleaning up after %T: %+v\n", factory, err)
		}
	}

	// return the test result
	os.Exit(result)
}

func runTest(t *testing.T, testFunc func(*testing.T, accounts.Storer, context.Context)) {
	t.Helper()

	logger := yall.New(colour.New(os.Stdout, yall.Debug))
	for _, factory := range factories {
		ctx := yall.InContext(context.Background(), logger)
		storer, err := factory.NewStorer(ctx)
		if err != nil {
			t.Fatalf("Error creating Storer from %T: %+v\n", factory, err)
		}
		t.Run(fmt.Sprintf("Storer=%T", storer), func(t *testing.T) {
			t.Parallel()
			testFunc(t, storer, ctx)
		})
	}
}

func TestCreateAndGetAccount(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuidOrFail(t),
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
		if diff := cmp.Diff(account, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}
	})
}

func TestGetNonexistentAccount(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		_, err := storer.Get(ctx, "myaccount@impractical.co")
		if !errors.Is(err, accounts.ErrAccountNotFound) {
			t.Fatalf("Expected ErrAccountNotFound, got %v\n", err)
		}
	})
}

func TestCreateDuplicateID(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuidOrFail(t),
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
			ProfileID: uuidOrFail(t),
			Created:   time.Now().Add(time.Hour).Round(time.Millisecond),
			LastUsed:  time.Now().Add(time.Hour).Round(time.Millisecond),
			LastSeen:  time.Now().Add(time.Hour).Round(time.Millisecond),
		}

		err = storer.Create(ctx, account2)
		if !errors.Is(err, accounts.ErrAccountAlreadyExists) {
			t.Fatalf("Expected ErrAccountAlreadyExists, got (%T) %s", err, err.Error())
		}

		// we shouldn't have changed anything about what was stored
		resp, err := storer.Get(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}

		if diff := cmp.Diff(account, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}
	})
}

func TestCreateSecondaryAccounts(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:             "paddy@impractical.co",
			ProfileID:      uuidOrFail(t),
			Created:        time.Now().Round(time.Millisecond),
			LastUsed:       time.Now().Round(time.Millisecond),
			LastSeen:       time.Now().Round(time.Millisecond),
			IsRegistration: true,
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}
		account2 := accounts.Account{
			ID:        "paddy@impracticallabs.com",
			ProfileID: account.ProfileID,
			Created:   time.Now().Add(time.Hour).Round(time.Millisecond),
			LastUsed:  time.Now().Add(time.Hour).Round(time.Millisecond),
			LastSeen:  time.Now().Add(time.Hour).Round(time.Millisecond),
		}

		err = storer.Create(ctx, account2)
		if err != nil {
			t.Fatalf("Unexpected error creating second account: %+v\n", err)
		}
		account3 := accounts.Account{
			ID:        "paddy@carvers.co",
			ProfileID: account.ProfileID,
			Created:   time.Now().Add(time.Hour).Round(time.Millisecond),
			LastUsed:  time.Now().Add(time.Hour).Round(time.Millisecond),
			LastSeen:  time.Now().Add(time.Hour).Round(time.Millisecond),
		}

		err = storer.Create(ctx, account3)
		if err != nil {
			t.Fatalf("Unexpected error creating third account: %+v\n", err)
		}

		resp, err := storer.Get(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}

		if diff := cmp.Diff(account, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}

		resp, err = storer.Get(ctx, account2.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}

		if diff := cmp.Diff(account2, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}

		resp, err = storer.Get(ctx, account3.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}

		if diff := cmp.Diff(account3, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}
	})
}

func TestCreateDuplicateRegistration(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:             "paddy@impractical.co",
			ProfileID:      uuidOrFail(t),
			Created:        time.Now().Round(time.Millisecond),
			LastUsed:       time.Now().Round(time.Millisecond),
			LastSeen:       time.Now().Round(time.Millisecond),
			IsRegistration: true,
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}
		account2 := accounts.Account{
			ID:             "paddy@impracticallabs.com",
			ProfileID:      account.ProfileID,
			Created:        time.Now().Add(time.Hour).Round(time.Millisecond),
			LastUsed:       time.Now().Add(time.Hour).Round(time.Millisecond),
			LastSeen:       time.Now().Add(time.Hour).Round(time.Millisecond),
			IsRegistration: true,
		}

		err = storer.Create(ctx, account2)
		if !errors.Is(err, accounts.ErrProfileIDAlreadyExists) {
			t.Fatalf("Expected ErrProfileIDAlreadyExists, got (%T) %v", err, err)
		}

		// we shouldn't have changed anything about what was stored
		resp, err := storer.Get(ctx, account.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}

		if diff := cmp.Diff(account, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}
	})
}

func TestCreateMultipleAccounts(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuidOrFail(t),
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
		if diff := cmp.Diff(account, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}

		resp, err = storer.Get(ctx, account2.ID)
		if err != nil {
			t.Fatalf("Unexpected error retrieving account: %+v\n", err)
		}
		if diff := cmp.Diff(account2, resp); diff != "" {
			t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
		}
	})
}

func TestListAccountsByProfile(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuidOrFail(t),
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
		if diff := cmp.Diff(account, accounts[0]); diff != "" {
			t.Errorf("Unexpected diff for %s (-wanted, +got): %s", account.ID, diff)
		}
		if diff := cmp.Diff(account2, accounts[1]); diff != "" {
			t.Errorf("Unexpected diff for %s (-wanted, +got): %s", account2.ID, diff)
		}
	})
}

func TestUpdateOneOfMany(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		for iter := 1; iter < changeVariations; iter++ {
			iter := iter
			t.Run(fmt.Sprintf("iter=%d", iter), func(t *testing.T) {
				t.Parallel()

				account := accounts.Account{
					ID:        fmt.Sprintf("paddy+%d@impractical.co", iter),
					ProfileID: uuidOrFail(t),
					Created:   time.Now().Round(time.Millisecond),
					LastUsed:  time.Now().Round(time.Millisecond),
					LastSeen:  time.Now().Round(time.Millisecond),
				}
				err := storer.Create(ctx, account)
				if err != nil {
					t.Fatalf("Unexpected error creating account: %+v\n", err)
				}

				var throwaways []accounts.Account
				for throwawayNum := 0; throwawayNum < 5; throwawayNum++ {
					throwaway := accounts.Account{
						ID:        fmt.Sprintf("paddy+%d+%d@impractical.co", iter, throwawayNum),
						ProfileID: uuidOrFail(t),
						Created:   time.Now().Add(time.Duration(throwawayNum) * time.Minute).Round(time.Millisecond),
						LastUsed:  time.Now().Add(time.Duration(throwawayNum) * time.Hour).Round(time.Millisecond),
						LastSeen:  time.Now().Add(time.Duration(throwawayNum) * time.Second).Round(time.Millisecond),
					}
					if throwawayNum%2 == 0 {
						throwaway.ProfileID = account.ProfileID
					}

					err = storer.Create(ctx, throwaway)
					if err != nil {
						t.Fatalf("Unexpected error creating account: %+v\n", err)
					}
					throwaways = append(throwaways, throwaway)
				}

				var change accounts.Change
				if iter&changeLastSeen != 0 {
					seen := time.Now().Add(time.Duration(iter) * time.Minute).Round(time.Millisecond)
					change.LastSeen = &seen
				}
				if iter&changeLastUsed != 0 {
					used := time.Now().Add(time.Duration(iter) * time.Hour).Round(time.Millisecond)
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
				if diff := cmp.Diff(expectation, result); diff != "" {
					t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
				}
				for _, throwaway := range throwaways {
					result, err := storer.Get(ctx, throwaway.ID)
					if err != nil {
						t.Errorf("Unexpected error retrieving account: %+v\n", err)
					}
					if diff := cmp.Diff(throwaway, result); diff != "" {
						t.Errorf("Unexpected diff for %s (-wanted, +got): %s", throwaway.ID, diff)
					}
				}
			})
		}
	})
}

func TestUpdateNonExistent(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		account := accounts.Account{
			ID:        "paddy@impractical.co",
			ProfileID: uuidOrFail(t),
			Created:   time.Now().Round(time.Millisecond),
			LastUsed:  time.Now().Round(time.Millisecond),
			LastSeen:  time.Now().Round(time.Millisecond),
		}
		err := storer.Create(ctx, account)
		if err != nil {
			t.Fatalf("Unexpected error creating account: %+v\n", err)
		}

		var throwaways []accounts.Account
		for throwawayNum := 0; throwawayNum < 5; throwawayNum++ {
			throwaway := accounts.Account{
				ID:        fmt.Sprintf("paddy+%d@impractical.co", throwawayNum),
				ProfileID: uuidOrFail(t),
				Created:   time.Now().Add(time.Duration(throwawayNum) * time.Minute).Round(time.Millisecond),
				LastUsed:  time.Now().Add(time.Duration(throwawayNum) * time.Hour).Round(time.Millisecond),
				LastSeen:  time.Now().Add(time.Duration(throwawayNum) * time.Second).Round(time.Millisecond),
			}
			if throwawayNum%2 == 0 {
				throwaway.ProfileID = account.ProfileID
			}

			err = storer.Create(ctx, throwaway)
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
		if !errors.Is(err, accounts.ErrAccountNotFound) {
			t.Logf("Account: %+v\n", res)
			t.Errorf("Expected error to be ErrAccountNotFound, got %v\n", err)
		}
		for _, throwaway := range throwaways {
			result, err := storer.Get(ctx, throwaway.ID)
			if err != nil {
				t.Errorf("Unexpected error retrieving account: %+v\n", err)
			}
			if diff := cmp.Diff(throwaway, result); diff != "" {
				t.Errorf("Unexpected diff for %s (-wanted, +got): %s", throwaway.ID, diff)
			}
		}
	})
}

func TestDeleteNonExistent(t *testing.T) {
	t.Parallel()

	runTest(t, func(t *testing.T, storer accounts.Storer, ctx context.Context) {
		// we shouldn't get an error deleting an account that doesn't exist
		err := storer.Delete(ctx, "notarealaccount@impractical.co")
		if err != nil {
			t.Fatalf("Unexpected error deleting account: %+v\n", err)
		}
	})
}

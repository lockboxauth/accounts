package apiv1

import (
	"net/http"

	"code.impractical.co/accounts"
	"darlinggo.co/api"
	"darlinggo.co/trout"
)

func (a APIv1) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var body Account
	err := api.Decode(r, &body)
	if err != nil {
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}
	account := coreAccount(body)
	account = accounts.FillDefaults(account)
	var reqErrs []api.RequestError
	if account.ID == "" {
		reqErrs = append(reqErrs, api.RequestError{Field: "/id", Slug: api.RequestErrMissing})
	}
	if account.ProfileID == "" {
		reqErrs = append(reqErrs, api.RequestError{Field: "/profileID", Slug: api.RequestErrMissing})
	}
	if len(reqErrs) > 0 {
		api.Encode(w, r, http.StatusBadRequest, reqErrs)
		return
	}
	err = a.Storer.Create(r.Context(), account)
	if err != nil {
		if err == accounts.ErrAccountAlreadyExists {
			api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Field: "/id", Slug: api.RequestErrConflict}}})
			return
		}
		a.Log.Printf("Error creating account: %+v\n", err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusCreated, Response{Accounts: []Account{apiAccount(account)}})
}

func (a APIv1) handleGetAccount(w http.ResponseWriter, r *http.Request) {
	vars := trout.RequestVars(r)
	id := vars.Get("id")
	if id == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "/id", Slug: api.RequestErrNotFound}}})
		return
	}
	account, err := a.Storer.Get(r.Context(), id)
	if err != nil {
		if err == accounts.ErrAccountNotFound {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Field: "/id", Slug: api.RequestErrNotFound}}})
			return
		}
		a.Log.Printf("Error retrieving account %s: %+v\n", id, err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Accounts: []Account{apiAccount(account)}})
}

func (a APIv1) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	vars := trout.RequestVars(r)
	id := vars.Get("id")
	if id == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "/id", Slug: api.RequestErrNotFound}}})
		return
	}
	var body Change
	err := api.Decode(r, &body)
	if err != nil {
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: api.InvalidFormatError})
		return
	}
	err = a.Storer.Update(r.Context(), id, coreChange(body))
	if err != nil {
		a.Log.Printf("Error updating account %s: %+v\n", id, err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	account, err := a.Storer.Get(r.Context(), id)
	if err != nil {
		if err == accounts.ErrAccountNotFound {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Field: "/id", Slug: api.RequestErrNotFound}}})
			return
		}
		a.Log.Printf("Error retrieving account %s: %+v\n", id, err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Accounts: []Account{apiAccount(account)}})
}

func (a APIv1) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := trout.RequestVars(r)
	id := vars.Get("id")
	if id == "" {
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
		return
	}
	account, err := a.Storer.Get(r.Context(), id)
	if err != nil {
		if err == accounts.ErrAccountNotFound {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Field: "/id", Slug: api.RequestErrNotFound}}})
			return
		}
		a.Log.Printf("Error retrieving account %s: %+v\n", id, err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	err = a.Storer.Delete(r.Context(), id)
	if err != nil {
		a.Log.Printf("Error deleting account %s: %+v\n", id, err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Accounts: []Account{apiAccount(account)}})
}

func (a APIv1) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	profileID := r.URL.Query().Get("profileID")
	if profileID == "" {
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Param: "profileID", Slug: api.RequestErrMissing}}})
		return
	}
	accts, err := a.Storer.ListByProfile(r.Context(), profileID)
	if err != nil {
		a.Log.Printf("Error retrieving accounts for %s: %+v\n", profileID, err)
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Accounts: apiAccounts(accts)})
}

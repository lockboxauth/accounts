package apiv1

import (
	"net/http"

	"darlinggo.co/api"
	"darlinggo.co/trout"
	"impractical.co/auth/accounts"
	yall "yall.in"
)

func (a APIv1) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var body Account
	err := api.Decode(r, &body)
	if err != nil {
		yall.FromContext(r.Context()).WithError(err).Debug("Error decoding request body")
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
		// TODO(paddy): if the profile is set, we need to make sure the request is authenticated by that profile or an admin scope
		// if this is set, that means we're adding an account to a profile, and we need auth to do that
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
		yall.FromContext(r.Context()).WithError(err).Error("Error creating account")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("account_id", account.ID).Debug("Account created")
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
		yall.FromContext(r.Context()).WithField("account_id", id).WithError(err).Error("Error retrieving account")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	// TODO(paddy): requester needs to either be the profile associated with the account or have an admin scope
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
		yall.FromContext(r.Context()).WithField("account_id", id).WithError(err).Error("Error retrieving account")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	// TODO(paddy): requester needs to either be the profile associated with the account or have an admin scope
	err = a.Storer.Delete(r.Context(), id)
	if err != nil {
		yall.FromContext(r.Context()).WithField("account_id", id).WithError(err).Error("Error deleting account")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	yall.FromContext(r.Context()).WithField("account_id", id).Debug("Account deleted")
	api.Encode(w, r, http.StatusOK, Response{Accounts: []Account{apiAccount(account)}})
}

func (a APIv1) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	profileID := r.URL.Query().Get("profileID")
	if profileID == "" {
		api.Encode(w, r, http.StatusBadRequest, Response{Errors: []api.RequestError{{Param: "profileID", Slug: api.RequestErrMissing}}})
		return
	}
	// TODO(paddy): requester needs to match the profile ID, or have an admin scope
	accts, err := a.Storer.ListByProfile(r.Context(), profileID)
	if err != nil {
		yall.FromContext(r.Context()).WithField("profile_id", profileID).WithError(err).Error("Error listing accounts")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Accounts: apiAccounts(accts)})
}

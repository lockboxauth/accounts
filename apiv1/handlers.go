package apiv1

import (
	"errors"
	"net/http"

	"darlinggo.co/api"
	"darlinggo.co/trout/v2"
	uuid "github.com/hashicorp/go-uuid"
	yall "yall.in"

	"lockbox.dev/accounts"
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
	if account.ProfileID == "" && !account.IsRegistration {
		reqErrs = append(reqErrs, api.RequestError{Field: "/profileID", Slug: api.RequestErrMissing})
	}
	if len(reqErrs) > 0 {
		api.Encode(w, r, http.StatusBadRequest, reqErrs)
		return
	}
	if account.ProfileID != "" {
		if resp := a.validateAddingAccountToProfile(r, account); resp != nil {
			api.Encode(w, r, resp.Status, resp)
			return
		}
	} else {
		var profileID string
		profileID, err = uuid.GenerateUUID()
		if err != nil {
			yall.FromContext(r.Context()).WithError(err).Error("error generating profile ID")
			api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
			return
		}
		account.ProfileID = profileID
	}
	err = a.Storer.Create(r.Context(), account)
	if err != nil {
		if errors.Is(err, accounts.ErrAccountAlreadyExists) {
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
		api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Param: "id", Slug: api.RequestErrNotFound}}})
		return
	}
	account, err := a.Storer.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, accounts.ErrAccountNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Field: "/id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithField("account_id", id).WithError(err).Error("Error retrieving account")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	sess, resp := a.GetAuthToken(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	if sess == nil {
		api.Encode(w, r, http.StatusUnauthorized, Response{Errors: []api.RequestError{
			{Header: "Authorization", Slug: api.RequestErrAccessDenied},
		}})
		return
	}
	if sess.ProfileID != account.ProfileID {
		api.Encode(w, r, http.StatusForbidden, Response{Errors: []api.RequestError{
			{Param: "id", Slug: api.RequestErrAccessDenied},
		}})
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
		if errors.Is(err, accounts.ErrAccountNotFound) {
			api.Encode(w, r, http.StatusNotFound, Response{Errors: []api.RequestError{{Field: "/id", Slug: api.RequestErrNotFound}}})
			return
		}
		yall.FromContext(r.Context()).WithField("account_id", id).WithError(err).Error("Error retrieving account")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	sess, resp := a.GetAuthToken(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	if sess == nil {
		api.Encode(w, r, http.StatusUnauthorized, Response{Errors: []api.RequestError{
			{Header: "Authorization", Slug: api.RequestErrAccessDenied},
		}})
		return
	}
	if sess.ProfileID != account.ProfileID {
		api.Encode(w, r, http.StatusForbidden, Response{Errors: []api.RequestError{
			{Param: "id", Slug: api.RequestErrAccessDenied},
		}})
		return
	}
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
	sess, resp := a.GetAuthToken(r)
	if resp != nil {
		api.Encode(w, r, resp.Status, resp)
		return
	}
	if sess == nil {
		api.Encode(w, r, http.StatusUnauthorized, Response{Errors: []api.RequestError{
			{Header: "Authorization", Slug: api.RequestErrAccessDenied},
		}})
		return
	}
	if sess.ProfileID != profileID {
		api.Encode(w, r, http.StatusForbidden, Response{Errors: []api.RequestError{
			{Param: "profileID", Slug: api.RequestErrAccessDenied},
		}})
		return
	}
	accts, err := a.Storer.ListByProfile(r.Context(), profileID)
	if err != nil {
		yall.FromContext(r.Context()).WithField("profile_id", profileID).WithError(err).Error("Error listing accounts")
		api.Encode(w, r, http.StatusInternalServerError, Response{Errors: api.ActOfGodError})
		return
	}
	api.Encode(w, r, http.StatusOK, Response{Accounts: apiAccounts(accts)})
}

func (a APIv1) validateAddingAccountToProfile(r *http.Request, account accounts.Account) *Response {
	sess, resp := a.GetAuthToken(r)
	if resp != nil {
		return nil
	}
	if sess == nil {
		return &Response{
			Status: http.StatusUnauthorized,
			Errors: []api.RequestError{
				{Header: "Authorization", Slug: api.RequestErrAccessDenied},
			},
		}
	}
	if sess.ProfileID != account.ProfileID {
		return &Response{
			Status: http.StatusForbidden,
			Errors: []api.RequestError{
				{Header: "Authorization", Slug: api.RequestErrAccessDenied},
			},
		}
	}
	return nil
}

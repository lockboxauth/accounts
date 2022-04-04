package apiv1

import (
	"context"
	"net/http"

	"darlinggo.co/trout/v2"
	"impractical.co/apidiags"
	"lockbox.dev/accounts"
	"lockbox.dev/sessions"
	yall "yall.in"
)

type deleteAccountHandler struct {
	a APIv1
}

type deleteAccountRequest struct {
	id   string
	sess *sessions.AccessToken
}

func (d deleteAccountHandler) parseRequest(ctx context.Context, r *http.Request, resp *Response) deleteAccountRequest {
	vars := trout.RequestVars(r)
	sess := d.a.GetAuthToken(r, resp)
	if resp.HasErrors() {
		return deleteAccountRequest{}
	}
	return deleteAccountRequest{
		id:   vars.Get("id"),
		sess: sess,
	}
}

func (d deleteAccountHandler) validateRequest(ctx context.Context, req deleteAccountRequest, resp *Response) {
	if req.id == "" {
		resp.SetStatus(http.StatusNotFound)
		resp.AddError(apidiags.CodeNotFound, apidiags.NewURLPointer("id"))
		return
	}
	if req.sess == nil {
		resp.SetStatus(http.StatusUnauthorized)
		resp.AddError(apidiags.CodeAccessDenied, apidiags.NewHeaderPointer("Authorization"))
		return
	}
}

func (d deleteAccountHandler) execute(ctx context.Context, req deleteAccountRequest, resp *Response) {
	account, err := d.a.Storer.Get(ctx, req.id)
	if err != nil {
		if err == accounts.ErrAccountNotFound {
			resp.SetStatus(http.StatusNotFound)
			resp.AddError(apidiags.CodeNotFound, apidiags.NewURLPointer("id"))
			return
		}
		yall.FromContext(ctx).WithField("account_id", req.id).WithError(err).Error("Error retrieving account")
		resp.SetStatus(http.StatusInternalServerError)
		resp.AddError(apidiags.CodeActOfGod)
		return
	}
	if req.sess.ProfileID != account.ProfileID {
		resp.SetStatus(http.StatusForbidden)
		resp.AddError(apidiags.CodeAccessDenied, apidiags.NewURLPointer("id"))
		return
	}
	err = d.a.Storer.Delete(ctx, req.id)
	if err != nil {
		yall.FromContext(ctx).WithField("account_id", req.id).WithError(err).Error("Error deleting account")
		resp.SetStatus(http.StatusInternalServerError)
		resp.AddError(apidiags.CodeActOfGod)
		return
	}
	yall.FromContext(ctx).WithField("account_id", req.id).Debug("Account deleted")
	resp.SetStatus(http.StatusOK)
	resp.AddAccounts(apiAccount(account))
}

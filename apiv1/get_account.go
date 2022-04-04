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

type getAccountHandler struct {
	a APIv1
}

type getAccountRequest struct {
	id   string
	sess *sessions.AccessToken
}

func (g getAccountHandler) parseRequest(ctx context.Context, r *http.Request, resp *Response) getAccountRequest {
	vars := trout.RequestVars(r)
	sess := g.a.GetAuthToken(r, resp)
	if resp.HasErrors() {
		return getAccountRequest{}
	}
	return getAccountRequest{
		id:   vars.Get("id"),
		sess: sess,
	}
}

func (g getAccountHandler) validateRequest(ctx context.Context, req getAccountRequest, resp *Response) {
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

func (g getAccountHandler) execute(ctx context.Context, req getAccountRequest, resp *Response) {
	account, err := g.a.Storer.Get(ctx, req.id)
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
	resp.SetStatus(http.StatusOK)
	resp.AddAccounts(apiAccount(account))
}

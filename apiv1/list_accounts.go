package apiv1

import (
	"context"
	"net/http"

	"impractical.co/apidiags"
	"lockbox.dev/sessions"
	yall "yall.in"
)

type listAccountsHandler struct {
	a APIv1
}

type listAccountsRequest struct {
	profileID string
	sess      *sessions.AccessToken
}

func (l listAccountsHandler) parseRequest(ctx context.Context, r *http.Request, resp *Response) listAccountsRequest {
	sess := l.a.GetAuthToken(r, resp)
	if resp.HasErrors() {
		return listAccountsRequest{}
	}
	return listAccountsRequest{
		profileID: r.URL.Query().Get("profileID"),
		sess:      sess,
	}
}

func (l listAccountsHandler) validateRequest(ctx context.Context, req listAccountsRequest, resp *Response) {
	if req.profileID == "" {
		resp.SetStatus(http.StatusBadRequest)
		resp.AddError(apidiags.CodeMissing, apidiags.NewURLPointer("profileID"))
		return
	}
	if req.sess == nil {
		resp.SetStatus(http.StatusUnauthorized)
		resp.AddError(apidiags.CodeAccessDenied, apidiags.NewHeaderPointer("Authorization"))
		return
	}
	if req.sess.ProfileID != req.profileID {
		resp.SetStatus(http.StatusForbidden)
		resp.AddError(apidiags.CodeAccessDenied, apidiags.NewURLPointer("profileID"))
		return
	}
}

func (l listAccountsHandler) execute(ctx context.Context, req listAccountsRequest, resp *Response) {
	accts, err := l.a.Storer.ListByProfile(ctx, req.profileID)
	if err != nil {
		yall.FromContext(ctx).WithField("profile_id", req.profileID).WithError(err).Error("Error listing accounts")
		resp.SetStatus(http.StatusInternalServerError)
		resp.AddError(apidiags.CodeActOfGod)
		return
	}
	resp.SetStatus(http.StatusOK)
	resp.AddAccounts(apiAccounts(accts)...)
}

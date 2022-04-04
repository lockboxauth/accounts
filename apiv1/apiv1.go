package apiv1

import (
	"net/http"

	"impractical.co/apidiags"
	yall "yall.in"

	"lockbox.dev/accounts"
	"lockbox.dev/sessions"
)

// APIv1 holds all the information that we want to
// be available for all the functions in the API,
// things like our logging, metrics, and other
// telemetry.
type APIv1 struct {
	accounts.Dependencies
	Log      *yall.Logger
	Sessions sessions.Dependencies
}

// GetAuthToken returns the access token associated
// with the request, or a Response that should be
// rendered if there's an error.
func (a APIv1) GetAuthToken(r *http.Request, resp *Response) *sessions.AccessToken {
	sess, err := a.Sessions.TokenFromRequest(r)
	if err != nil {
		if err == sessions.ErrInvalidToken {
			resp.SetStatus(http.StatusUnauthorized)
			resp.AddError(apidiags.CodeAccessDenied, apidiags.NewHeaderPointer("Authorization"))
			return nil
		}
		yall.FromContext(r.Context()).WithError(err).Error("Error decoding session")
		resp.SetStatus(http.StatusInternalServerError)
		resp.AddError(apidiags.CodeActOfGod)
		return nil
	}
	return sess
}

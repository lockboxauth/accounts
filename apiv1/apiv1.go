package apiv1

import (
	"net/http"

	"darlinggo.co/api"
	"impractical.co/auth/accounts"
	"impractical.co/auth/sessions"
	yall "yall.in"
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

func (a APIv1) GetAuthToken(r *http.Request) (*sessions.AccessToken, *Response) {
	sess, err := a.Sessions.TokenFromRequest(r)
	if err != nil {
		if err == sessions.ErrInvalidToken {
			return nil, &Response{
				Errors: []api.RequestError{{
					Header: "Authorization",
					Slug:   api.RequestErrAccessDenied,
				}},
				Status: http.StatusUnauthorized,
			}
		} else {
			yall.FromContext(r.Context()).WithError(err).Error("Error decoding session")
			return nil, &Response{
				Errors: api.ActOfGodError,
				Status: http.StatusInternalServerError,
			}
		}
	}
	return sess, nil
}

// Response is used to encode JSON responses; it is
// the global response format for all API responses.
type Response struct {
	Accounts []Account          `json:"accounts,omitempty"`
	Errors   []api.RequestError `json:"errors,omitempty"`
	Status   int                `json:"-"`
}

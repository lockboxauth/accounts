package apiv1

import (
	"darlinggo.co/api"
	"impractical.co/auth/accounts"
	yall "yall.in"
)

// APIv1 holds all the information that we want to
// be available for all the functions in the API,
// things like our logging, metrics, and other
// telemetry.
type APIv1 struct {
	accounts.Dependencies
	Log *yall.Logger
}

// Response is used to encode JSON responses; it is
// the global response format for all API responses.
type Response struct {
	Accounts []Account          `json:"accounts,omitempty"`
	Errors   []api.RequestError `json:"errors,omitempty"`
}

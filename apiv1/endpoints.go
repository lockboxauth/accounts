package apiv1

import (
	"net/http"

	"darlinggo.co/api"
	"darlinggo.co/trout"
)

// Server returns an http.Handler that will handle all
// the requests for v1 of the API. The baseURL should be
// set to whatever prefix the muxer matches to pass requests
// to the Handler; consider it the root path of v1 of the API.
func (a APIv1) Server(baseURL string) http.Handler {
	var router trout.Router
	router.SetPrefix(baseURL)
	router.Endpoint("/").Methods("POST").Handler(api.NegotiateMiddleware(http.HandlerFunc(a.handleCreateAccount)))
	router.Endpoint("/").Methods("GET").Handler(api.NegotiateMiddleware(http.HandlerFunc(a.handleListAccounts)))
	router.Endpoint("/{id}").Methods("GET").Handler(api.NegotiateMiddleware(http.HandlerFunc(a.handleGetAccount)))
	router.Endpoint("/{id}").Methods("DELETE").Handler(api.NegotiateMiddleware(http.HandlerFunc(a.handleDeleteAccount)))

	return router
}

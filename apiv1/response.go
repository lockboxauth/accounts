package apiv1

import (
	"context"
	"net/http"

	"github.com/adjust/goautoneg"
	"github.com/google/go-cmp/cmp"
	"impractical.co/apidiags"
	yall "yall.in"
)

var encoders = []encoder{
	jsonEncoder{},
}

// Response represents information that should be conveyed back to the client.
type Response struct {
	enc    encoder `json:"-"`
	status int     `json:"-"`

	Accounts []Account             `json:"accounts,omitempty"`
	Diags    []apidiags.Diagnostic `json:"diags,omitempty"`
}

// Equal returns true if two Responses should be considered equal. It is
// largely used to make testing Responses using go-cmp easier.
func (r Response) Equal(o Response) bool {
	if r.enc != o.enc {
		return false
	}
	if r.status != o.status {
		return false
	}
	if !cmp.Equal(r.Accounts, o.Accounts) {
		return false
	}
	if !cmp.Equal(r.Diags, o.Diags) {
		return false
	}
	return true
}

// newResponse returns a Response that is ready to be used, priming it to
// encode data in a format that matches the Accept header of the request.
func newResponse(ctx context.Context, r *http.Request) *Response {
	var resp Response
	alts := make([]string, 0, len(encoders))
	encTypes := map[string]encoder{}
	for _, enc := range encoders {
		alts = append(alts, enc.contentType())
		encTypes[enc.contentType()] = enc
	}
	resp.enc = encTypes[goautoneg.Negotiate(r.Header.Get("Accept"), alts)]
	if resp.enc == nil {
		resp.enc = jsonEncoder{}
		yall.FromContext(ctx).WithField("accept", r.Header.Get("Accept")).Warn("no known encoder matched Accept header, defaulting to JSON encoder")
	}
	return &resp
}

// HasErrors returns true if the Response has any error level diagnostics.
func (r *Response) HasErrors() bool {
	for _, diag := range r.Diags {
		if diag.Severity == apidiags.DiagnosticError {
			return true
		}
	}
	return false
}

// Send encodes the Response in a format that matches the Accept header, if at
// all possible, and writes it to the passed http.ResponseWriter.
func (r *Response) Send(ctx context.Context, w http.ResponseWriter) {
	body, err := r.enc.encode(*r)
	if err != nil {
		yall.FromContext(ctx).WithError(err).Error("error encoding response")
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write([]byte("An unexpected error occurred"))
		if writeErr != nil {
			yall.FromContext(ctx).WithError(writeErr).Error("error writing encoding error response")
		}
		return
	}
	w.Header().Set("Content-Type", r.enc.contentType())
	w.WriteHeader(r.status)
	n, err := w.Write(body)
	if err != nil {
		yall.FromContext(ctx).WithError(err).Error("error writing response")
		return
	}
	if n != len(body) {
		yall.FromContext(ctx).WithField("encoded", len(body)).
			WithField("sent", n).
			Warn("a different number of bytes were encoded than were sent")
		return
	}
}

// SetStatus sets the HTTP status code of the Response. It cannot be called
// after Response.Send.
func (r *Response) SetStatus(status int) {
	r.status = status
}

// AddError appends an error-level diagnostic to the Response. It cannot be
// called after Response.Send.
func (r *Response) AddError(code apidiags.Code, pointers ...apidiags.Pointer) {
	r.Diags = append(r.Diags, apidiags.Diagnostic{
		Severity: apidiags.DiagnosticError,
		Code:     code,
		Pointers: pointers,
	})
}

// AddWarning appends a warning-level diagnostic to the Response. It cannot be
// called after Response.Send.
func (r *Response) AddWarning(code apidiags.Code, pointers ...apidiags.Pointer) {
	r.Diags = append(r.Diags, apidiags.Diagnostic{
		Severity: apidiags.DiagnosticWarning,
		Code:     code,
		Pointers: pointers,
	})
}

// AddAccounts appends one or more Accounts to the Response. It cannot be
// called after Response.Send.
func (r *Response) AddAccounts(accounts ...Account) {
	r.Accounts = append(r.Accounts, accounts...)
}

// An encoder is a strategy for converting a Response into bytes in response to
// an Accept header.
type encoder interface {
	encode(Response) ([]byte, error)
	contentType() string
}

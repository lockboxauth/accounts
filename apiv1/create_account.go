package apiv1

import (
	"context"
	"net/http"

	"darlinggo.co/api"
	uuid "github.com/hashicorp/go-uuid"
	"impractical.co/apidiags"
	"lockbox.dev/accounts"
	yall "yall.in"
)

type createAccountHandler struct {
	a APIv1
}

func (c createAccountHandler) parseRequest(ctx context.Context, r *http.Request, resp *Response) Account {
	var body Account
	err := api.Decode(r, &body)
	if err != nil {
		yall.FromContext(ctx).WithError(err).Debug("Error decoding request body")
		resp.SetStatus(http.StatusBadRequest)
		resp.AddError(apidiags.CodeInvalidFormat, apidiags.NewBodyPointer())
		return Account{}
	}
	acc := coreAccount(body)
	acc = accounts.FillDefaults(acc)
	account := apiAccount(acc)
	if account.ProfileID != "" {
		sess := c.a.GetAuthToken(r, resp)
		if resp.HasErrors() {
			return Account{}
		}
		if sess == nil {
			resp.SetStatus(http.StatusUnauthorized)
			resp.AddError(apidiags.CodeAccessDenied, apidiags.NewHeaderPointer("Authorization"))
			return Account{}
		}
		if sess.ProfileID != account.ProfileID {
			resp.SetStatus(http.StatusForbidden)
			resp.AddError(apidiags.CodeAccessDenied, apidiags.NewHeaderPointer("Authorization"))
			return Account{}
		}
	} else {
		profileID, err := uuid.GenerateUUID()
		if err != nil {
			yall.FromContext(ctx).WithError(err).Error("error generating profile ID")
			resp.SetStatus(http.StatusInternalServerError)
			resp.AddError(apidiags.CodeActOfGod)
			return Account{}
		}
		account.ProfileID = profileID
	}
	return account
}

func (c createAccountHandler) validateRequest(ctx context.Context, req Account, resp *Response) {
	if req.ID == "" {
		resp.SetStatus(http.StatusBadRequest)
		resp.AddError(apidiags.CodeMissing, apidiags.NewBodyPointer("id"))
	}
	if req.ProfileID == "" && !req.IsRegistration {
		resp.SetStatus(http.StatusBadRequest)
		resp.AddError(apidiags.CodeMissing, apidiags.NewBodyPointer("profileID"))
	}
	if resp.HasErrors() {
		return
	}
	return
}

func (c createAccountHandler) execute(ctx context.Context, req Account, resp *Response) {
	err := c.a.Storer.Create(ctx, coreAccount(req))
	if err != nil {
		if err == accounts.ErrAccountAlreadyExists {
			resp.SetStatus(http.StatusBadRequest)
			resp.AddError(apidiags.CodeConflict, apidiags.NewBodyPointer("id"))
			return
		}
		yall.FromContext(ctx).WithError(err).Error("Error creating account")
		resp.SetStatus(http.StatusInternalServerError)
		resp.AddError(apidiags.CodeActOfGod)
		return
	}
	yall.FromContext(ctx).WithField("account_id", req.ID).Debug("Account created")
	resp.SetStatus(http.StatusCreated)
	resp.AddAccounts(req)
}

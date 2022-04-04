package apiv1

import (
	"context"
	"net/http"

	"darlinggo.co/api"
	"impractical.co/apidiags"
	yall "yall.in"
)

type handler[RequestType any] interface {
	execute(context.Context, RequestType, *Response)
}

type handlerWithValidate[RequestType any] interface {
	validateRequest(context.Context, RequestType, *Response)
}

type handlerWithParse[RequestType any] interface {
	handler[RequestType]

	parseRequest(context.Context, *http.Request, *Response) RequestType
}

func defaultRequestParser[RequestType any](ctx context.Context, r *http.Request, resp *Response) RequestType {
	var req RequestType
	err := api.Decode(r, &req)
	if err != nil {
		yall.FromContext(ctx).WithError(err).Debug("Error decoding request body")
		resp.SetStatus(http.StatusBadRequest)
		resp.AddError(apidiags.CodeInvalidFormat, apidiags.NewBodyPointer())
	}
	return req
}

func handle[RequestType any](h handler[RequestType]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RequestType
		resp := newResponse(r.Context(), r)
		if parser, ok := h.(handlerWithParse[RequestType]); ok {
			req = parser.parseRequest(r.Context(), r, resp)
		} else {
			req = defaultRequestParser[RequestType](r.Context(), r, resp)
		}
		if resp.HasErrors() {
			resp.Send(r.Context(), w)
			return
		}

		if validator, ok := h.(handlerWithValidate[RequestType]); ok {
			validator.validateRequest(r.Context(), req, resp)
			if resp.HasErrors() {
				resp.Send(r.Context(), w)
				return
			}
		}

		h.execute(r.Context(), req, resp)
		resp.Send(r.Context(), w)
	})
}

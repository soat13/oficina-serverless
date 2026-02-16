package apigw

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

func Router(ctx context.Context, h *Handler, req events.APIGatewayV2HTTPRequest) events.APIGatewayV2HTTPResponse {
	method := req.RequestContext.HTTP.Method
	path := req.RawPath

	switch {
	case method == http.MethodPost && path == "/auth/token":
		return h.PostToken(ctx, req)
	case method == http.MethodPost && path == "/auth/verify":
		return h.PostVerify(ctx, req)
	default:
		return jsonError(http.StatusNotFound, "not_found")
	}
}

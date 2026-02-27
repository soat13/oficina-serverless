package apigw

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

func Router(ctx context.Context, h *Handler, req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	return h.PostToken(ctx, req)
}

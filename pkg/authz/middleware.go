package authz

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/jsmithdenverdev/pager/pkg/problemdetails"
)

type (
	apiGatewayLambdaHandler    = func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
	apiGatewayLambdaMiddleware = func(next apiGatewayLambdaHandler) apiGatewayLambdaHandler
)

func APIGatewayWithClient(verifiedPermissionsClient *verifiedpermissions.Client) apiGatewayLambdaMiddleware {
	return func(next apiGatewayLambdaHandler) apiGatewayLambdaHandler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			var userInfo map[string]interface{}
			if err := json.Unmarshal([]byte(event.Headers["x-pager-userinfo"]), &userInfo); err != nil {
				return problemdetails.MarshalToAPIGatewayProxyResponse(problemdetails.ProblemDetails{
					Type:   "auth/authorization",
					Title:  "Unauthorized",
					Status: http.StatusUnauthorized,
				})
			}

			entitlementAttributeValues := []types.AttributeValue{}

			userEntitlements, ok := userInfo["entitlements"].([]interface{})
			if !ok {
				return problemdetails.MarshalToAPIGatewayProxyResponse(problemdetails.ProblemDetails{
					Type:   "auth/authorization",
					Title:  "Unauthorized",
					Status: http.StatusUnauthorized,
				})
			}

			for _, userEntitlement := range userEntitlements {
				entitlementAttributeValues = append(entitlementAttributeValues, &types.AttributeValueMemberString{
					Value: userEntitlement.(string),
				})
			}

			client := NewClient(
				WithVerifiedPermissionsClient(verifiedPermissionsClient),
				WithUserInfo(UserInfo{}))

			ctx = context.WithValue(ctx, contextKeyClient, client)

			return next(ctx, event)
		}
	}
}

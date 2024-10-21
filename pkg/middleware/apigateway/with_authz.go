package apigateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/middleware"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
)

func WithAuthz(verifiedPermissionsClient *verifiedpermissions.Client) middleware.APIGatewayLambdaMiddleware {
	return func(next middleware.APIGatewayLambdaHandler) middleware.APIGatewayLambdaHandler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			var userInfo map[string]interface{}
			var resp events.APIGatewayProxyResponse

			if err := json.Unmarshal([]byte(event.Headers["x-pager-userinfo"]), &userInfo); err != nil {
				if err := problemdetail.WriteToAPIGatewayProxyResponse(
					&resp,
					problemdetail.New(
						"auth/authorization",
						problemdetail.WithTitle("Unauthorized")),
					http.StatusUnauthorized); err != nil {
					writeInternalServerError(&resp)
				}

				return resp, err
			}

			entitlementAttributeValues := []types.AttributeValue{}

			userEntitlements, ok := userInfo["entitlements"].([]interface{})
			if !ok {
				if err := problemdetail.WriteToAPIGatewayProxyResponse(
					&resp,
					problemdetail.New(
						"auth/authorization",
						problemdetail.WithTitle("Unauthorized")),
					http.StatusUnauthorized); err != nil {
					writeInternalServerError(&resp)
				}
				return resp, errors.New("unable to convert entitlements to array")
			}

			for _, userEntitlement := range userEntitlements {
				entitlementAttributeValues = append(entitlementAttributeValues, &types.AttributeValueMemberString{
					Value: userEntitlement.(string),
				})
			}

			client := authz.NewClient(
				authz.WithVerifiedPermissionsClient(verifiedPermissionsClient),
				authz.WithUserInfo(authz.UserInfo{}))

			ctx = authz.AddClientToContext(ctx, client)

			return next(ctx, event)
		}
	}
}

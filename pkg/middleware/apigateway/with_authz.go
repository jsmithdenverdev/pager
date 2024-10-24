package apigateway

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	"github.com/jsmithdenverdev/pager/pkg/authz"
	"github.com/jsmithdenverdev/pager/pkg/codec/apigateway"
	"github.com/jsmithdenverdev/pager/pkg/middleware"
	"github.com/jsmithdenverdev/pager/pkg/problemdetail"
)

func WithAuthz(policyStoreId string, verifiedPermissionsClient *verifiedpermissions.Client, logger *slog.Logger) middleware.APIGatewayLambdaMiddleware {
	return func(next middleware.APIGatewayLambdaHandler) middleware.APIGatewayLambdaHandler {
		return func(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			var (
				userInfo  authz.UserInfo
				pdEncoder = apigateway.NewEncoder[problemdetail.ProblemDetailer]()
			)

			if err := json.Unmarshal([]byte(event.Headers["x-pager-userinfo"]), &userInfo); err != nil {
				logger.ErrorContext(
					ctx,
					"failed to decode user info from header",
					slog.Any("decode error", err))

				resp, encErr := pdEncoder.Encode(
					ctx,
					problemdetail.New("internal-server-error"))

				if encErr != nil {
					logger.ErrorContext(
						ctx,
						"failed to encode response",
						slog.Any("encode error", encErr))
				}

				return resp, nil
			}

			if agencyId, ok := event.PathParameters["agencyid"]; ok {
				userInfo.ActiveAgency = agencyId
			}

			logger.DebugContext(
				ctx,
				"authz middleware",
				slog.Any("userInfo", userInfo))

			client := authz.NewClient(
				authz.WithVerifiedPermissionsClient(verifiedPermissionsClient),
				authz.WithUserInfo(userInfo),
				authz.WithPolicyStoreID(policyStoreId))

			ctx = authz.AddClientToContext(ctx, client)
			ctx = authz.AddUserInfoToContext(ctx, userInfo)

			return next(ctx, event)
		}
	}
}

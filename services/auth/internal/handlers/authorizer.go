package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jsmithdenverdev/pager/services/auth/internal/config"
)

func Authorizer(config config.Config, logger *slog.Logger, client *dynamodb.Client) func(context.Context, events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	type userInfoResponse struct {
		ID     string `json:"id"`
		Email  string `json:"email"`
		IDPID  string `json:"idpId"`
		Status string `json:"status"`
	}

	return func(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
		logger.InfoContext(ctx, "request for authorization", slog.Any("event", event))

		tokenString := getTokenFromHeader(event.Headers["authorization"])
		if tokenString == "" {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", "no token in header"))
			return events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "Anonymous",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Deny",
							Resource: []string{"*"},
						},
					},
				},
			}, nil
		}

		// Fetch the JWKS from Auth0
		jwks, err := keyfunc.NewDefault([]string{fmt.Sprintf("https://%s/.well-known/jwks.json", config.Auth0Domain)})
		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "Anonymous",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Deny",
							Resource: []string{"*"},
						},
					},
				},
			}, nil
		}

		// Parse and validate the JWT
		claims, err := verifyToken(tokenString, "", "", jwks)
		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "Anonymous",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Deny",
							Resource: []string{"*"},
						},
					},
				},
			}, nil
		}

		// Ensure audience matches
		sub := claims["sub"].(string)

		row, err := client.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(config.TableName),
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{
					Value: sub,
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "Anonymous",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Deny",
							Resource: []string{"*"},
						},
					},
				},
			}, nil
		}

		userJSON, err := json.Marshal(row.Item)

		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return events.APIGatewayCustomAuthorizerResponse{
				PrincipalID: "Anonymous",
				PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
					Version: "2012-10-17",
					Statement: []events.IAMPolicyStatement{
						{
							Action:   []string{"execute-api:Invoke"},
							Effect:   "Deny",
							Resource: []string{"*"},
						},
					},
				},
			}, nil
		}

		return events.APIGatewayCustomAuthorizerResponse{
			PrincipalID: sub,
			PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
				Version: "2012-10-17",
				Statement: []events.IAMPolicyStatement{
					{
						Action:   []string{"execute-api:Invoke"},
						Effect:   "Allow",
						Resource: []string{event.MethodArn},
					},
				},
			},
			Context: map[string]interface{}{
				"userid":   sub,
				"userinfo": string(userJSON),
			},
		}, nil

	}
}

func getTokenFromHeader(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}

// verifyToken verifies the JWT using the JWKS and returns the claims if valid
func verifyToken(tokenString string, audience string, auth0Domain string, jwks keyfunc.Keyfunc) (jwt.MapClaims, error) {
	token, err := jwt.Parse(
		tokenString,
		jwks.Keyfunc,
		jwt.WithAudience(audience),
		jwt.WithIssuer(auth0Domain))

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to parse claims")
	}

	return claims, nil
}

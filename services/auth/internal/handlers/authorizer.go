package handlers

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jsmithdenverdev/pager/services/auth/internal/config"
	"github.com/lestrrat-go/jwx/jwk"
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
		jwks, err := jwk.Fetch(ctx, fmt.Sprintf("https://%s/.well-known/jwks.json", config.Auth0Domain))
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
		claims, err := verifyToken(tokenString, config.Auth0Audience, config.Auth0Domain, jwks)
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
						Action: []string{"execute-api:Invoke"},
						Effect: "Allow",
						// TODO: Find out why method arn is not propagating
						Resource: []string{event.MethodArn, "*"},
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
func verifyToken(tokenString string, audience string, auth0Domain string, jwks jwk.Set) (jwt.MapClaims, error) {
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("no kid present in the token header")
		}

		// Find the key in the JWKS that matches the kid
		key, found := jwks.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("unable to find key with kid: %s", kid)
		}

		var pubkey rsa.PublicKey
		if err := key.Raw(&pubkey); err != nil {
			return nil, fmt.Errorf("unable to parse RSA public key: %w", err)
		}

		return &pubkey, nil
	})
	if err != nil {
		return jwt.MapClaims{}, fmt.Errorf("error parsing token: %w", err)
	}

	// Validate token claims (audience and issuer)
	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		err := validateClaims(claims, auth0Domain, audience)
		if err != nil {
			return jwt.MapClaims{}, err
		}
		return claims, nil
	} else {
		return jwt.MapClaims{}, errors.New("invalid token")
	}
}

func validateClaims(claims jwt.MapClaims, auth0Domain, audience string) error {
	iss := fmt.Sprintf("https://%s/", auth0Domain)
	if claims["iss"] != iss {
		return fmt.Errorf("invalid issuer: %v", claims["iss"])
	}

	audClaimSlice, ok := claims["aud"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid audience: %s", "could not convert aud to interface slice")
	}

	var audMatch bool
	for _, aud := range audClaimSlice {
		if aud, ok := aud.(string); ok {
			if aud == audience {
				audMatch = true
				break
			}
		}

	}

	if !audMatch {
		return fmt.Errorf("invalid audience: %v", claims["aud"])
	}

	return nil
}

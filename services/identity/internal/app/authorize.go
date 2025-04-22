package app

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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/lestrrat-go/jwx/jwk"
)

func Authorize(config Config, logger *slog.Logger, client *dynamodb.Client) func(context.Context, events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	return func(ctx context.Context, request events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
		logger.InfoContext(ctx, "authorizing request", slog.Any("request", request))

		response := events.APIGatewayCustomAuthorizerResponse{
			PrincipalID: "Anonymous",
			PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
				Version: "2012-10-17",
				Statement: []events.IAMPolicyStatement{
					{
						Effect:   "Deny",
						Resource: []string{"*"},
						Action:   []string{"execute-api:Invoke"},
					},
				},
			},
		}

		token := getTokenFromHeader(request.Headers["Authorization"])
		if token == "" {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", "no token in header"))
			return response, nil
		}

		jwks, err := jwk.Fetch(ctx, fmt.Sprintf("https://%s/.well-known/jwks.json", config.Auth0Domain))
		if err != nil {
			logger.ErrorContext(ctx, "failed to fetch jwks", slog.String("error", err.Error()))
			return response, nil
		}

		claims, err := verifyToken(token, config.Auth0Audience, config.Auth0Domain, jwks)
		if err != nil {
			logger.ErrorContext(ctx, "failed to verify token", slog.String("error", err.Error()))
			return response, nil
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			logger.ErrorContext(ctx, "failed to get sub claim", slog.String("error", "sub claim is not a string"))
			return response, nil
		}

		userRecord, err := client.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(config.TableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("idpid#%s", sub),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("idpid#%s", sub),
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "failed to get user record", slog.String("error", err.Error()))
			return response, nil
		}

		if userRecord.Item == nil {
			logger.ErrorContext(ctx, "user not found", slog.String("sub", sub))
			return response, nil
		}

		var user identity.User
		if err := attributevalue.UnmarshalMap(userRecord.Item, &user); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal user record", slog.String("error", err.Error()))
			return response, nil
		}

		userJSON, err := json.Marshal(user)
		if err != nil {
			logger.ErrorContext(ctx, "failed to marshal user record", slog.String("error", err.Error()))
			return response, nil
		}

		response.PrincipalID = sub
		response.PolicyDocument.Statement[0].Effect = "Allow"
		response.Context = map[string]interface{}{
			"userid": sub,
			"user":   string(userJSON),
		}

		return response, nil
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

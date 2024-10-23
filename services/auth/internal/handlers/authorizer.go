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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jsmithdenverdev/pager/services/auth/internal/config"
	"github.com/lestrrat-go/jwx/jwk"
)

type userInfo struct {
	ID           string   `dynamodbav:"id" json:"id"`
	Email        string   `dynamodbav:"email" json:"email"`
	IDPID        string   `dynamodbav:"idpId" json:"idpId"`
	Status       string   `dynamodbav:"status" json:"status"`
	Entitlements []string `dynamodbav:"entitlements" json:"entitlements"`
	Agencies     []struct {
		Roles []string `dynamodbav:"roles" json:"roles"`
	} `dynamodbav:"agencies" json:"agencies"`
}

func Authorizer(config config.Config, logger *slog.Logger, client *dynamodb.Client) func(context.Context, events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	return func(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
		// Create a deny response by default. If auth succeeds we modify the
		// response to an Approve and add the userid and userinfo to context.
		response := events.APIGatewayCustomAuthorizerResponse{
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
		}

		// Fetch token from header
		tokenString := getTokenFromHeader(event.Headers["authorization"])
		if tokenString == "" {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", "no token in header"))
			return response, nil
		}

		// Fetch the JWKS from Auth0
		jwks, err := jwk.Fetch(ctx, fmt.Sprintf("https://%s/.well-known/jwks.json", config.Auth0Domain))
		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return response, nil
		}

		// Parse and validate the JWT
		claims, err := verifyToken(tokenString, config.Auth0Audience, config.Auth0Domain, jwks)
		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return response, nil
		}

		// Fetch sub (user id) from token claims
		sub := claims["sub"].(string)
		// Add the sub as the principal ID in the response policy
		response.PrincipalID = sub

		// Using the sub, fetch the user details for this user.
		result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(config.TableName),
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{
					Value: sub,
				},
			},
		})

		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return response, nil
		}

		if result.Item == nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", "user not found"), slog.String("id", sub))
			return response, nil
		}

		// Unmarshal the user dynamodb record into a userInfo struct
		var userInfo userInfo
		err = attributevalue.UnmarshalMap(result.Item, &userInfo)

		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return response, nil
		}

		// marhsal userInfo into json
		userJSON, err := json.Marshal(userInfo)

		if err != nil {
			logger.ErrorContext(ctx, "authorization failed", slog.String("error", err.Error()))
			return response, nil
		}

		// We've succesfully validated the token and fetched the user, convert the
		// policy effect to ALlow, add userid and userinfo to the response context.
		response.PolicyDocument.Statement[0].Effect = "Allow"
		response.Context = map[string]interface{}{
			"userid":   sub,
			"userinfo": string(userJSON),
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

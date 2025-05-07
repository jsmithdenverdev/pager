package app

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
	"github.com/jsmithdenverdev/pager/pkg/identity"
)

func createEndpoint(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user        identity.User
			userinfostr = r.Header.Get("x-pager-userinfo")
		)

		if err := json.Unmarshal([]byte(userinfostr), &user); err != nil {
			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Platform Admins cannot create endpoints for themselves, they do not
		// belong to an agency
		if slices.Contains(user.Entitlements, identity.EntitlementPlatformAdmin) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// A user can only create an endpoint if they have a membership in an agency
		// TODO: what about agency status?
		if len(user.Memberships) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			if err := json.NewEncoder(w).Encode(map[string]string{
				"error": "you must have at least one agency membership to create an endpoint",
			}); err != nil {
				logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		req, problems, err := decodeValid[createEndpointRequest](r)
		if err != nil {
			if len(problems) > 0 {
				w.WriteHeader(http.StatusBadRequest)
				if err := json.NewEncoder(w).Encode(problems); err != nil {
					logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		now := time.Now()
		id := uuid.New().String()

		dynamoInput, err := attributevalue.MarshalMap(endpoint{
			keyFields: keyFields{
				PK:   fmt.Sprintf("endpoint#%s", id),
				SK:   "meta",
				Type: entityTypeEndpoint,
			},
			auditableFields: newAuditableFields(user.ID, now),
			UserID:          user.ID,
			Name:            req.Name,
			EndpointType:    req.EndpointType,
			URL:             req.URL,
		})
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal endpoint", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		registrationCodeDynamoInput, err := attributevalue.MarshalMap(registrationCode{
			keyFields: keyFields{
				PK:   fmt.Sprintf("rc#%x", sha256.Sum256([]byte(id))),
				SK:   "registrationcode",
				Type: entityTypeRegistrationCode,
			},
			auditableFields: newAuditableFields(user.ID, now),
			EndpointID:      id,
			UserID:          user.ID,
		})
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal registration code", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ownershipLinkDynamoInput, err := attributevalue.MarshalMap(ownershipLink{
			keyFields: keyFields{
				PK:   fmt.Sprintf("user#%s", user.ID),
				SK:   fmt.Sprintf("endpoint#%s", id),
				Type: entityTypeOwnershipLink,
			},
			auditableFields: newAuditableFields(user.ID, now),
		})
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal ownership link", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Use DynamoDB TransactWriteItems for atomic creation
		_, err = dynamoClient.TransactWriteItems(r.Context(), &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Put: &types.Put{
						TableName: aws.String(config.EndpointTableName),
						Item:      dynamoInput,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String(config.EndpointTableName),
						Item:      registrationCodeDynamoInput,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String(config.EndpointTableName),
						Item:      ownershipLinkDynamoInput,
					},
				},
			},
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to transact write endpoint entities", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err = encode(w, r, int(http.StatusCreated), createEndpointResponse{ID: id}); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

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
	"github.com/jsmithdenverdev/pager/services/endpoint/internal/models"
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

		var (
			now              = time.Now()
			id               = uuid.New().String()
			registrationCode = fmt.Sprintf("%x", sha256.Sum256([]byte(id)))
		)

		endpointAV, err := attributevalue.MarshalMap(models.Endpoint{
			KeyFields: models.KeyFields{
				PK:   fmt.Sprintf("endpoint#%s", id),
				SK:   "meta",
				Type: models.EntityTypeEndpoint,
			},
			AuditableFields:  models.NewAuditableFields(user.ID, now),
			UserID:           user.ID,
			Name:             req.Name,
			EndpointType:     req.EndpointType,
			RegistrationCode: registrationCode,
			URL:              req.URL,
		})
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal endpoint", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		registrationCodeAV, err := attributevalue.MarshalMap(models.RegistrationCode{
			KeyFields: models.KeyFields{
				PK:   fmt.Sprintf("rc#%s", registrationCode),
				SK:   "registrationcode",
				Type: models.EntityTypeRegistrationCode,
			},
			AuditableFields: models.NewAuditableFields(user.ID, now),
			EndpointID:      id,
			UserID:          user.ID,
		})
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal registration code", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ownershipLinkAV, err := attributevalue.MarshalMap(models.OwnershipLink{
			KeyFields: models.KeyFields{
				PK:   fmt.Sprintf("user#%s", user.ID),
				SK:   fmt.Sprintf("endpoint#%s", id),
				Type: models.EntityTypeOwnershipLink,
			},
			AuditableFields: models.NewAuditableFields(user.ID, now),
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
						Item:      endpointAV,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String(config.EndpointTableName),
						Item:      registrationCodeAV,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String(config.EndpointTableName),
						Item:      ownershipLinkAV,
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

package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/jsmithdenverdev/pager/pkg/identity"
)

// createAgency creates a new agency.
func createAgency(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
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

		if !slices.Contains(user.Entitlements, identity.EntitlementPlatformAdmin) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		req, problems, err := decodeValid[createAgencyRequest](r)
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

		id := uuid.New().String()

		dynamoInput, err := attributevalue.MarshalMap(agency{
			PK:         fmt.Sprintf("agency#%s", id),
			SK:         fmt.Sprintf("agency#%s", id),
			Type:       entityTypeAgency,
			Name:       req.Name,
			Status:     agencyStatusActive,
			Created:    time.Now(),
			Modified:   time.Now(),
			CreatedBy:  user.ID,
			ModifiedBy: user.ID,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal agency", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = client.PutItem(r.Context(), &dynamodb.PutItemInput{
			TableName: aws.String(config.AgencyTableName),
			Item:      dynamoInput,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to put agency", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err = encode(w, r, int(http.StatusCreated), createAgencyResponse{ID: id}); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

func registerEndpoint(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user     identity.User
			agencyID = r.PathValue("id")
		)

		if err := json.Unmarshal([]byte(r.Header.Get("x-pager-userinfo")), &user); err != nil {
			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if role, ok := user.Memberships[agencyID]; !ok || role != identity.RoleWriter {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		req, problems, err := decodeValid[registerEndpointRequest](r)
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

		registrationAV, err := attributevalue.MarshalMap(models.EndpointRegistration{
			PK:         fmt.Sprintf("agency#%s", agencyID),
			SK:         fmt.Sprintf("registration#%s", req.RegistrationCode),
			Type:       models.EntityTypeRegistration,
			Status:     models.RegistrationStatusPending,
			Created:    now,
			Modified:   now,
			CreatedBy:  user.ID,
			ModifiedBy: user.ID,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to write registration", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = dynamoClient.PutItem(r.Context(), &dynamodb.PutItemInput{
			TableName: aws.String(config.AgencyTableName),
			Item:      registrationAV,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to write registration", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		messageBody, err := json.Marshal(struct {
			AgencyID         string `json:"agencyId"`
			RegistrationCode string `json:"registrationCode"`
		}{
			AgencyID:         agencyID,
			RegistrationCode: req.RegistrationCode,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal SNS message", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = snsClient.Publish(r.Context(), &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(string(messageBody)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("endpoint.resolve"),
				},
			},
		}); err != nil {
			logger.ErrorContext(r.Context(), "failed to publish SNS message", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = encode(w, r, http.StatusCreated, registerEndpointResponse{
			RegistrationCode: req.RegistrationCode,
			AgencyID:         agencyID,
			Status:           models.RegistrationStatusPending,
			Created:          now,
			Modified:         now,
			CreatedBy:        user.ID,
			ModifiedBy:       user.ID,
		}); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

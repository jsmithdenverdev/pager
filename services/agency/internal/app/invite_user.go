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
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/jsmithdenverdev/pager/pkg/identity"
)

func inviteUser(config Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) http.Handler {
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

		req, problems, err := decodeValid[createMembershipRequest](r)
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

		userAgencyMembershipAV, err := attributevalue.MarshalMap(membership{
			PK:         fmt.Sprintf("user#%s", req.UserID),
			SK:         fmt.Sprintf("agency#%s", agencyID),
			Type:       entityTypeMembership,
			Role:       req.Role,
			Status:     membershipStatusPending,
			Created:    now,
			Modified:   now,
			CreatedBy:  user.ID,
			ModifiedBy: user.ID,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal user agency membership", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		agencyUserMembershipAV, err := attributevalue.MarshalMap(membership{
			PK:         fmt.Sprintf("agency#%s", agencyID),
			SK:         fmt.Sprintf("user#%s", req.UserID),
			Type:       entityTypeMembership,
			Role:       req.Role,
			Status:     membershipStatusPending,
			Created:    now,
			Modified:   now,
			CreatedBy:  user.ID,
			ModifiedBy: user.ID,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal agency user membership", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		dynamoInput := &dynamodb.TransactWriteItemsInput{
			TransactItems: []dynamotypes.TransactWriteItem{
				{
					Put: &dynamotypes.Put{
						TableName: aws.String(config.AgencyTableName),
						Item:      userAgencyMembershipAV,
					},
				},
				{
					Put: &dynamotypes.Put{
						TableName: aws.String(config.AgencyTableName),
						Item:      agencyUserMembershipAV,
					},
				},
			},
		}

		_, err = dynamoClient.TransactWriteItems(r.Context(), dynamoInput)

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to write memberships", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = snsClient.Publish(r.Context(), &sns.PublishInput{
			TopicArn: aws.String(config.EventsTopicARN),
			Message:  aws.String(fmt.Sprintf("User %s invited to agency %s", req.UserID, agencyID)),
			MessageAttributes: map[string]snstypes.MessageAttributeValue{
				"type": {
					DataType:    aws.String("String"),
					StringValue: aws.String("membership.invite"),
				},
			},
		}); err != nil {
			logger.ErrorContext(r.Context(), "failed to publish SNS message", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err = encode(w, r, http.StatusCreated, createMembershipResponse{
			AgencyID:   agencyID,
			UserID:     req.UserID,
			Role:       req.Role,
			Status:     membershipStatusPending,
			Created:    now,
			Modified:   now,
			CreatedBy:  user.ID,
			ModifiedBy: user.ID,
		}); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jsmithdenverdev/pager/pkg/identity"
)

// readAgency returns a single agency by ID.
// The calling user must have a membership in the specified agency.
func readAgency(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user        identity.User
			userinfostr = r.Header.Get("x-pager-userinfo")
			agencyid    = r.PathValue("id")
		)

		if err := json.Unmarshal([]byte(userinfostr), &user); err != nil {
			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, ok := user.Memberships[agencyid]; !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		result, err := client.GetItem(r.Context(), &dynamodb.GetItemInput{
			TableName: aws.String(config.AgencyTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", agencyid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: "meta",
				},
			},
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to get agency", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var agency agency
		if result.Item != nil {
			if err := attributevalue.UnmarshalMap(result.Item, &agency); err != nil {
				logger.ErrorContext(r.Context(), "failed to unmarshal agency record", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if err := json.NewEncoder(w).Encode(toAgencyResponse(agency)); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

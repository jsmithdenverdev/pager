package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

// listMemberships returns a list of memberships in the specified agency.
// The calling user must have a membership in the specified agency.
func listMemberships(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err         error
			user        identity.User
			first       = 10
			firstStr    = r.URL.Query().Get("first")
			cursor      = r.URL.Query().Get("cursor")
			userinfostr = r.Header.Get("x-pager-userinfo")
			agencyid    = r.PathValue("id")
		)

		if firstStr != "" {
			first, err = strconv.Atoi(firstStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if err := json.Unmarshal([]byte(userinfostr), &user); err != nil {
			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, ok := user.Memberships[agencyid]; !ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(config.AgencyTableName),
			Limit:                  aws.Int32(int32(first)),
			KeyConditionExpression: aws.String("#pk = :pk AND begins_with(#sk, :sk)"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", agencyid),
				},
				":sk": &types.AttributeValueMemberS{Value: "user#"},
			},
		}

		if cursor != "" {
			queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", agencyid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", cursor),
				},
			}
		}

		result, err := client.Query(r.Context(), queryInput)

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to query agencies", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var memberships []models.Membership
		if result.Items != nil {
			for _, item := range result.Items {
				var membership models.Membership
				if err := attributevalue.UnmarshalMap(item, &membership); err != nil {
					logger.ErrorContext(r.Context(), "failed to unmarshal membership record", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				memberships = append(memberships, membership)
			}
		}

		response := new(listResponse[membershipResponse])

		for _, membership := range memberships {
			response.Results = append(response.Results, membershipResponse{
				AgencyID: strings.Split(membership.PK, "#")[1],
				UserID:   strings.Split(membership.SK, "#")[1],
				Role:     membership.Role,
			})
		}

		if result.LastEvaluatedKey != nil {
			response.NextCursor = strings.Split(result.LastEvaluatedKey["sk"].(*types.AttributeValueMemberS).Value, "#")[1]
			response.HasNextPage = true
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
)

// listAgencies returns a list of agencies the calling user is a member of.
func listAgencies(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			user     identity.User
			first    = 10
			userid   = r.Header.Get("x-pager-userid")
			firstStr = r.URL.Query().Get("first")
			cursor   = r.URL.Query().Get("cursor")
		)

		if err := json.Unmarshal([]byte(r.Header.Get("x-pager-userinfo")), &user); err != nil {
			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if firstStr != "" {
			first, err = strconv.Atoi(firstStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		if slices.Contains(user.Entitlements, identity.EntitlementPlatformAdmin) {
			scanInput := dynamodb.ScanInput{
				TableName:        aws.String(config.AgencyTableName),
				Limit:            aws.Int32(int32(first)),
				FilterExpression: aws.String("begins_with(#pk, :pkprefix) AND #sk = :skprefix"),
				ExpressionAttributeNames: map[string]string{
					"#pk": "pk",
					"#sk": "sk",
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":pkprefix": &types.AttributeValueMemberS{Value: "agency#"},
					":skprefix": &types.AttributeValueMemberS{Value: "meta"},
				},
			}

			if cursor != "" {
				scanInput.ExclusiveStartKey = map[string]types.AttributeValue{
					"pk": &types.AttributeValueMemberS{
						Value: fmt.Sprintf("agency#%s", cursor),
					},
					"sk": &types.AttributeValueMemberS{
						Value: "meta",
					},
				}
			}

			result, err := client.Scan(r.Context(), &scanInput)
			if err != nil {
				logger.ErrorContext(r.Context(), "failed to scan agencies", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			var agencies []models.Agency
			if result.Items != nil {
				for _, item := range result.Items {
					var agency models.Agency
					if err := attributevalue.UnmarshalMap(item, &agency); err != nil {
						logger.ErrorContext(r.Context(), "failed to unmarshal agency record", slog.Any("error", err))
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					agencies = append(agencies, agency)
				}
			}

			response := new(listResponse[agencyResponse])

			for _, agency := range agencies {
				response.Results = append(response.Results, toAgencyResponse(agency))
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

			return
		}

		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(config.AgencyTableName),
			Limit:                  aws.Int32(int32(first)),
			KeyConditionExpression: aws.String("#pk = :userid AND begins_with(#sk, :skprefix)"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":userid": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", userid),
				},
				":skprefix": &types.AttributeValueMemberS{
					Value: "agency#",
				},
			},
		}

		if cursor != "" {
			queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", userid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", cursor),
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
				AgencyID: strings.Split(membership.SK, "#")[1],
				UserID:   strings.Split(membership.PK, "#")[1],
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

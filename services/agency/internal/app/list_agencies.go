package app

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func listMemberships(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	type membershipResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Role string `json:"role"`
	}

	type listMembershipsResponse struct {
		Memberships []membershipResponse `json:"agencies"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idpid := r.Header.Get("x-pager-userid")

		result, err := client.Query(r.Context(), &dynamodb.QueryInput{
			TableName:              &config.AgencyTableName,
			KeyConditionExpression: aws.String("pk = :idpid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":idpid": &types.AttributeValueMemberS{
					Value: idpid,
				},
			},
		})
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to query agencies", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var memberships []membership
		if result.Items != nil {
			for _, item := range result.Items {
				var membership membership
				if err := attributevalue.UnmarshalMap(item, &membership); err != nil {
					logger.ErrorContext(r.Context(), "failed to unmarshal membership record", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				memberships = append(memberships, membership)
			}
		}

		response := new(listMembershipsResponse)

		for _, membership := range memberships {
			id := strings.Split(membership.PK, "#")[1]
			response.Memberships = append(response.Memberships, membershipResponse{
				ID:   id,
				Name: membership.Name,
				Role: membership.Role,
			})
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

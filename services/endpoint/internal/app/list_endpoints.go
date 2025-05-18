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
	"github.com/jsmithdenverdev/pager/services/endpoint/internal/models"
)

// listEndpoints returns a list of endpoints. If an account ID is provided the
// endpoints registered to the account are returned. Otherwise the endpoints
// registered to the calling user are returned.
func listEndpoints(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			first    = 10
			userid   = r.Header.Get("x-pager-userid")
			firstStr = r.URL.Query().Get("first")
			cursor   = r.URL.Query().Get("cursor")
		)

		if firstStr != "" {
			first, err = strconv.Atoi(firstStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(config.EndpointTableName),
			Limit:                  aws.Int32(int32(first)),
			KeyConditionExpression: aws.String("#pk = :pk AND begins_with(#sk, :sk)"),
			ExpressionAttributeNames: map[string]string{
				"#pk": "pk",
				"#sk": "sk",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", userid),
				},
				":sk": &types.AttributeValueMemberS{
					Value: "endpoint#",
				},
			},
		}

		if cursor != "" {
			queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", userid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("endpoint#%s", cursor),
				},
			}
		}

		result, err := client.Query(r.Context(), queryInput)
		if err != nil {
			logger.ErrorContext(r.Context(), "failed to query endpoints", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var links []models.OwnershipLink
		if result.Items != nil {
			for _, item := range result.Items {
				var link models.OwnershipLink
				if err := attributevalue.UnmarshalMap(item, &link); err != nil {
					logger.ErrorContext(r.Context(), "failed to unmarshal ownership link record", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				links = append(links, link)
			}
		}

		response := new(listResponse[ownershipLinkResponse])

		for _, link := range links {
			response.Results = append(response.Results, toOwnershipLinkResponse(link))
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

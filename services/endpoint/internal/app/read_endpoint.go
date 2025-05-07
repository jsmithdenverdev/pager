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
)

// readEndpoint returns a single endpoint by ID.
// The calling user must have a membership in the specified agency.
func readEndpoint(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			userid     = r.Header.Get("x-pager-userid")
			endpointid = r.PathValue("id")
		)

		result, err := client.GetItem(r.Context(), &dynamodb.GetItemInput{
			TableName: aws.String(config.EndpointTableName),
			Key: map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", userid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("endpoint#%s", endpointid),
				},
			},
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to get endpoint", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var endpoint endpoint
		if result.Item != nil {
			if err := attributevalue.UnmarshalMap(result.Item, &endpoint); err != nil {
				logger.ErrorContext(r.Context(), "failed to unmarshal endpoint record", slog.Any("error", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if err := json.NewEncoder(w).Encode(endpoint); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

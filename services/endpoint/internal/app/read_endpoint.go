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
					Value: fmt.Sprintf("endpoint#%s", endpointid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: "meta",
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

		// Ensure the endpoint belongs to the calling user. This means we're doing
		// a read no matter what which isn't ideal, but it's the only way to
		// enforce this.
		// We could also store the endpoint using the userid as the primary key,
		// which would force us to only return the endpoint if it belongs to the
		// user, but that breaks away from the Unauthorized pattern we use for
		// other reads.
		if endpoint.UserID != userid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := json.NewEncoder(w).Encode(endpoint); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

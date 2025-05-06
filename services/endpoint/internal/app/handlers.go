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
			KeyConditionExpression: aws.String("pk = :userid AND begins_with(sk, :skprefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":userid": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("user#%s", userid),
				},
				":skprefix": &types.AttributeValueMemberS{Value: "endpoint#"},
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

		var endpoints []endpoint
		if result.Items != nil {
			for _, item := range result.Items {
				var endpoint endpoint
				if err := attributevalue.UnmarshalMap(item, &endpoint); err != nil {
					logger.ErrorContext(r.Context(), "failed to unmarshal endpoint record", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				endpoints = append(endpoints, endpoint)
			}
		}

		response := new(listResponse[endpointResponse])

		for _, endpoint := range endpoints {
			response.Results = append(response.Results, endpointResponse{
				ID:           strings.Split(endpoint.SK, "#")[1],
				UserID:       strings.Split(endpoint.PK, "#")[1],
				EndpointType: endpoint.EndpointType,
				Name:         endpoint.Name,
				URL:          endpoint.URL,
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

// // listRegistrations returns a list of registrations for the specified agency.
// func listRegistrations(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var (
// 			err         error
// 			user        identity.User
// 			first       = 10
// 			userinfostr = r.Header.Get("x-pager-userinfo")
// 			firstStr    = r.URL.Query().Get("first")
// 			cursor      = r.URL.Query().Get("cursor")
// 			agencyid    = r.PathValue("agencyid")
// 		)

// 		if firstStr != "" {
// 			first, err = strconv.Atoi(firstStr)
// 			if err != nil {
// 				w.WriteHeader(http.StatusBadRequest)
// 				return
// 			}
// 		}

// 		if err := json.Unmarshal([]byte(userinfostr), &user); err != nil {
// 			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		if _, ok := user.Memberships[agencyid]; !ok {
// 			w.WriteHeader(http.StatusForbidden)
// 			return
// 		}

// 		queryInput := &dynamodb.QueryInput{
// 			TableName:              aws.String(config.EndpointTableName),
// 			Limit:                  aws.Int32(int32(first)),
// 			KeyConditionExpression: aws.String("pk = :agencyid AND begins_with(sk, :skprefix)"),
// 			ExpressionAttributeValues: map[string]types.AttributeValue{
// 				":agencyid": &types.AttributeValueMemberS{
// 					Value: fmt.Sprintf("agency#%s", agencyid),
// 				},
// 				":skprefix": &types.AttributeValueMemberS{Value: "endpoint#"},
// 			},
// 		}

// 		if cursor != "" {
// 			queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
// 				"pk": &types.AttributeValueMemberS{
// 					Value: fmt.Sprintf("agency#%s", agencyid),
// 				},
// 				"sk": &types.AttributeValueMemberS{
// 					Value: fmt.Sprintf("endpoint#%s", cursor),
// 				},
// 			}
// 		}

// 		result, err := client.Query(r.Context(), queryInput)
// 		if err != nil {
// 			logger.ErrorContext(r.Context(), "failed to query endpoints", slog.Any("error", err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}

// 		var registrations []registration
// 		if result.Items != nil {
// 			for _, item := range result.Items {
// 				var registration registration
// 				if err := attributevalue.UnmarshalMap(item, &registration); err != nil {
// 					logger.ErrorContext(r.Context(), "failed to unmarshal registration record", slog.Any("error", err))
// 					w.WriteHeader(http.StatusInternalServerError)
// 					return
// 				}
// 				registrations = append(registrations, registration)
// 			}
// 		}

// 		response := new(listResponse[registrationResponse])

// 		for _, registration := range registrations {
// 			response.Results = append(response.Results, registrationResponse{
// 				AccountID:  strings.Split(registration.PK, "#")[1],
// 				EndpointID: strings.Split(registration.SK, "#")[1],
// 				UserID:     registration.UserID,
// 			})
// 		}

// 		if result.LastEvaluatedKey != nil {
// 			response.NextCursor = strings.Split(result.LastEvaluatedKey["sk"].(*types.AttributeValueMemberS).Value, "#")[1]
// 			response.HasNextPage = true
// 		}

// 		if err := json.NewEncoder(w).Encode(response); err != nil {
// 			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 	})
// }

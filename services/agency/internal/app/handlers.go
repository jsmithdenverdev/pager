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
)

// listMyMemberships returns a list of agencies the calling user is a member of.
func listMyMemberships(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			err      error
			first    = 10
			idpid    = r.Header.Get("x-pager-userid")
			firstStr = r.URL.Query().Get("first")
			cursor   = r.URL.Query().Get("cursor")
		)

		logger.InfoContext(r.Context(), "listMyMemberships", slog.Any("url query", r.URL.Query()))

		if firstStr != "" {
			first, err = strconv.Atoi(firstStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		queryInput := &dynamodb.QueryInput{
			TableName:              aws.String(config.AgencyTableName),
			Limit:                  aws.Int32(int32(first)),
			KeyConditionExpression: aws.String("pk = :idpid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":idpid": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("idpid#%s", idpid),
				},
			},
		}

		if cursor != "" {
			queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("idpid#%s", idpid),
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

// listAgencyMemberships returns a list of memberships in the specified agency.
// The calling user must have a membership in the specified agency.
func listAgencyMemberships(config Config, logger *slog.Logger, client *dynamodb.Client) http.Handler {
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
			KeyConditionExpression: aws.String("pk = :agencyid AND begins_with(sk, :skprefix)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":agencyid": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", agencyid),
				},
				":skprefix": &types.AttributeValueMemberS{Value: "idpid#"},
			},
		}

		if cursor != "" {
			queryInput.ExclusiveStartKey = map[string]types.AttributeValue{
				"pk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("agency#%s", agencyid),
				},
				"sk": &types.AttributeValueMemberS{
					Value: fmt.Sprintf("idpid#%s", cursor),
				},
			}
		}

		result, err := client.Query(r.Context(), queryInput)

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

		logger.InfoContext(
			r.Context(),
			"readAgency",
			slog.Any("url query", r.URL.Query()),
			slog.Any("user info", user))

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
					Value: fmt.Sprintf("agency#%s", agencyid),
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

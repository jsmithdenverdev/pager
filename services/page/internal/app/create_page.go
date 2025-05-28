package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/page/internal/models"
)

func createPage(conf Config, logger *slog.Logger, dynamoClient *dynamodb.Client, snsClient *sns.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			user        identity.User
			userinfostr = r.Header.Get("x-pager-userinfo")
		)

		if err := json.Unmarshal([]byte(userinfostr), &user); err != nil {
			logger.ErrorContext(r.Context(), "failed to unmarshal user info", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req, problems, err := decodeValid[createPageRequest](r)
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

		// Check all the agencies the user is attempting to send the page to. If
		// they are not a member of all agencies return a forbidden response.
		for _, agency := range req.Agencies {
			if _, ok := user.Memberships[agency]; !ok {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		id := uuid.New().String()

		dynamoInput, err := attributevalue.MarshalMap(models.Page{
			PK:    fmt.Sprintf("page#%s", id),
			SK:    "meta",
			Type:  models.EntityTypePage,
			Title: req.Title,
			Notes: req.Notes,
			Location: models.Location{
				CommonName: req.Location.CommonName,
				Latitude:   req.Location.Latitude,
				Longitude:  req.Location.Longitude,
				Type:       req.Location.Type,
			},
			Created:    time.Now(),
			Modified:   time.Now(),
			CreatedBy:  user.ID,
			ModifiedBy: user.ID,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to marshal page", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = dynamoClient.PutItem(r.Context(), &dynamodb.PutItemInput{
			TableName: aws.String(conf.PageTableName),
			Item:      dynamoInput,
		})

		if err != nil {
			logger.ErrorContext(r.Context(), "failed to put page", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.Notify {
			for _, agency := range req.Agencies {
				messageBody, err := json.Marshal(struct {
					Title    string `json:"title"`
					PageID   string `json:"pageId"`
					AgencyID string `json:"agencyId"`
				}{
					Title:    req.Title,
					PageID:   id,
					AgencyID: agency,
				})

				if err != nil {
					logger.ErrorContext(r.Context(), "failed to marshal SNS message", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				if _, err = snsClient.Publish(r.Context(), &sns.PublishInput{
					TopicArn: aws.String(conf.EventsTopicARN),
					Message:  aws.String(string(messageBody)),
					MessageAttributes: map[string]snstypes.MessageAttributeValue{
						"type": {
							DataType:    aws.String("String"),
							StringValue: aws.String("endpoint.deliver-page"),
						},
					},
				}); err != nil {
					logger.ErrorContext(r.Context(), "failed publish to SNS", slog.Any("error", err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		w.WriteHeader(http.StatusCreated)
		if err = encode(w, r, int(http.StatusCreated), createPageResponse{ID: id}); err != nil {
			logger.ErrorContext(r.Context(), "failed to encode response", slog.Any("error", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

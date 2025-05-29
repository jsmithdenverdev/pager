package app

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/google/uuid"
	"github.com/jsmithdenverdev/pager/pkg/identity"
	"github.com/jsmithdenverdev/pager/services/agency/internal/models"
	"log/slog"
	"time"
)

// service is a request scoped agency service.
type service struct {
	user         identity.User
	config       Config
	logger       *slog.Logger
	dynamoClient *dynamodb.Client
	snsClient    *sns.Client
}

func (s *service) createAgency(ctx context.Context, name string) (string, error) {
	id := uuid.New().String()

	dynamoInput, err := attributevalue.MarshalMap(models.Agency{
		PK:         fmt.Sprintf("agency#%s", id),
		SK:         "meta",
		Type:       models.EntityTypeAgency,
		Name:       name,
		Status:     models.AgencyStatusActive,
		Created:    time.Now(),
		Modified:   time.Now(),
		CreatedBy:  s.user.ID,
		ModifiedBy: s.user.ID,
	})

	if err != nil {
		return "", fmt.Errorf("failed to marshal agency: %w", err)
	}
	_, err = s.dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.config.AgencyTableName),
		Item:      dynamoInput,
	})

	if err != nil {
		return "", fmt.Errorf("failed to put agency: %w", err)
	}

	return id, nil
}

package attributevalue_test

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jsmithdenverdev/pager/pkg/attributevalue"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

type user struct {
	ID    string `dynamodbav:"id"`
	Name  string `dynamodbav:"name"`
	Email string `dynamodbav:"email"`
}

func (u user) Type() string {
	return "USER"
}

func (u user) EncodeKey() attributevalue.Key {
	return attributevalue.Key{
		PK: fmt.Sprintf("user#%s", u.ID),
		SK: "meta",
	}
}

func (u user) DecodeKey(key attributevalue.Key) error {
	pkSplit := strings.Split(key.PK, "#")
	if len(pkSplit) != 2 {
		return fmt.Errorf("invalid PK: %s", key.PK)
	}
	u.ID = pkSplit[1]
	return nil
}

func TestMarshalMap(t *testing.T) {
	u := user{
		ID:    "123",
		Name:  "Jon Doe",
		Email: "fake@.com",
	}

	// The expected output now includes the row structure
	expected := map[string]types.AttributeValue{
		"pk":   &types.AttributeValueMemberS{Value: "user#123"},
		"sk":   &types.AttributeValueMemberS{Value: "meta"},
		"type": &types.AttributeValueMemberS{Value: "USER"},
		"row": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"id":    &types.AttributeValueMemberS{Value: "123"},
			"name":  &types.AttributeValueMemberS{Value: "Jon Doe"},
			"email": &types.AttributeValueMemberS{Value: "fake@.com"},
		}},
	}

	actual, err := attributevalue.MarshalMap(u)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, actual)
}

func TestUnmarshalMap(t *testing.T) {
	// Create a map that represents a DynamoDB item with the row structure
	item := map[string]types.AttributeValue{
		"pk":   &types.AttributeValueMemberS{Value: "user#123"},
		"sk":   &types.AttributeValueMemberS{Value: "meta"},
		"type": &types.AttributeValueMemberS{Value: "USER"},
		"row": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"id":    &types.AttributeValueMemberS{Value: "123"},
			"name":  &types.AttributeValueMemberS{Value: "Jon Doe"},
			"email": &types.AttributeValueMemberS{Value: "fake@.com"},
		}},
	}

	// Create a user to unmarshal into
	u := user{
		// Initialize with some values to ensure they get overwritten
		ID:    "initial",
		Name:  "initial",
		Email: "initial",
	}

	// Unmarshal the item into the user
	err := attributevalue.UnmarshalMap(item, &u)

	// Check for errors
	assert.NoError(t, err)

	// Verify the user was properly unmarshaled
	assert.Equal(t, "123", u.ID)
	assert.Equal(t, "Jon Doe", u.Name)
	assert.Equal(t, "fake@.com", u.Email)
}

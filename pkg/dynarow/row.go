package dynarow

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type row struct {
	RowBuilder
}

func (r row) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	type alias struct {
		Row RowBuilder `dynamodbav:"row"`
		Key
		Type string `dynamodbav:"type"`
	}
	a := alias{
		Row: r.RowBuilder,
	}
	a.Key = a.Row.EncodeKey()
	a.Type = a.Row.Type()
	item, err := attributevalue.MarshalMap(a)
	if err != nil {
		return nil, err
	}
	return &types.AttributeValueMemberM{Value: item}, nil
}

func (r row) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	mm, ok := av.(*types.AttributeValueMemberM)
	if !ok {
		return fmt.Errorf("expected map attribute, got %T", av)
	}
	item := mm.Value

	// Retrieve and unmarshal pk
	pkAttr, ok := item["pk"]
	if !ok {
		return errors.New("missing pk")
	}
	var pk string
	if err := attributevalue.Unmarshal(pkAttr, &pk); err != nil {
		return fmt.Errorf("failed to unmarshal pk: %w", err)
	}

	// Retrieve and unmarshal sk
	skAttr, ok := item["sk"]
	if !ok {
		return errors.New("missing sk")
	}
	var sk string
	if err := attributevalue.Unmarshal(skAttr, &sk); err != nil {
		return fmt.Errorf("failed to unmarshal sk: %w", err)
	}

	// We can't unmarshal directly into the RowBuilder interface,
	// so we'll just pass the key to DecodeKey below

	// Decode pk and sk into model
	if err := r.RowBuilder.DecodeKey(Key{
		PK: pk,
		SK: sk,
	}); err != nil {
		return fmt.Errorf("failed to decode key: %w", err)
	}

	return nil
}

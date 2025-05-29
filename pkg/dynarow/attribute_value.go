package dynarow

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Key struct {
	PK string `dynamodbav:"pk"`
	SK string `dynamodbav:"sk"`
}

type RowBuilder interface {
	EncodeKey() Key
	DecodeKey(Key) error
	Type() string
}

func MarshalMap[T RowBuilder](rb T) (map[string]types.AttributeValue, error) {
	return attributevalue.MarshalMap(row{
		RowBuilder: rb,
	})
}

func UnmarshalMap[T RowBuilder](m map[string]types.AttributeValue, rb T) error {
	// Extract the key fields from the map
	pkAttr, ok := m["pk"]
	if !ok {
		return fmt.Errorf("missing pk")
	}
	var pk string
	if err := attributevalue.Unmarshal(pkAttr, &pk); err != nil {
		return fmt.Errorf("failed to unmarshal pk: %w", err)
	}

	skAttr, ok := m["sk"]
	if !ok {
		return fmt.Errorf("missing sk")
	}
	var sk string
	if err := attributevalue.Unmarshal(skAttr, &sk); err != nil {
		return fmt.Errorf("failed to unmarshal sk: %w", err)
	}

	// Extract the row data
	rowAttr, ok := m["row"]
	if !ok {
		return fmt.Errorf("missing row")
	}

	// Unmarshal the row data directly into rb
	if err := attributevalue.Unmarshal(rowAttr, rb); err != nil {
		return fmt.Errorf("failed to unmarshal row: %w", err)
	}

	// Set the key fields
	if err := (rb).DecodeKey(Key{PK: pk, SK: sk}); err != nil {
		return fmt.Errorf("failed to decode key: %w", err)
	}

	return nil
}

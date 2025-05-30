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
	// First marshal the row builder to get all its fields
	rbMap, err := attributevalue.MarshalMap(rb)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal row builder: %w", err)
	}

	// Get the key and type
	key := rb.EncodeKey()
	typeVal := rb.Type()

	// Marshal the key to get pk and sk
	keyMap, err := attributevalue.MarshalMap(key)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	// Marshal the type
	typeAttr, err := attributevalue.Marshal(typeVal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal type: %w", err)
	}

	// Create the result map with all fields at the top level
	result := rbMap

	// Add the key fields (pk, sk)
	for k, v := range keyMap {
		result[k] = v
	}

	// Add the type field
	result["type"] = typeAttr

	return result, nil
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

	// Create a copy of the map without the key and type fields
	// to avoid unmarshaling them into the row builder
	flatMap := make(map[string]types.AttributeValue)
	for k, v := range m {
		if k != "pk" && k != "sk" && k != "type" {
			flatMap[k] = v
		}
	}

	// Unmarshal the flat map directly into rb
	if err := attributevalue.UnmarshalMap(flatMap, rb); err != nil {
		return fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	// Set the key fields
	if err := (rb).DecodeKey(Key{PK: pk, SK: sk}); err != nil {
		return fmt.Errorf("failed to decode key: %w", err)
	}

	return nil
}

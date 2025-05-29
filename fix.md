# Fix for UnmarshalMap Panic in attributevalue Package

## Issue Description

The `UnmarshalMap` function in the custom attributevalue package was causing a panic with the error message:

```
reflect: reflect.Value.SetString using unaddressable value
```

This error occurs when the Go reflection system tries to set a value on a non-addressable (non-pointer) value.

## Root Cause Analysis

The issue was in the implementation of the `UnmarshalMap` function in `attribute_value.go`:

```go
func UnmarshalMap[T RowBuilder](m map[string]types.AttributeValue, rb *T) error {
    return attributevalue.UnmarshalMap(m, &row{
        RowBuilder: *rb,
    })
}
```

The problem is that:

1. The function creates a new `row` struct with the dereferenced `RowBuilder` value (`*rb`).
2. It then passes a pointer to this new `row` struct to `attributevalue.UnmarshalMap`.
3. Inside the AWS SDK's `attributevalue.UnmarshalMap`, it tries to set values on the `row` struct.
4. The `RowBuilder` field in the `row` struct is an interface value, not a pointer, so when the AWS SDK tries to set values on it using reflection, it gets the "unaddressable value" panic.

The issue is that we're trying to unmarshal into a non-addressable value (the interface field).

## Solution

The fix involves changing how we handle unmarshaling in both the `UnmarshalMap` function and the `UnmarshalDynamoDBAttributeValue` method of the `row` struct.

### Changes to `UnmarshalMap` function

```go
func UnmarshalMap[T RowBuilder](m map[string]types.AttributeValue, rb *T) error {
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
    if err := (*rb).DecodeKey(Key{PK: pk, SK: sk}); err != nil {
        return fmt.Errorf("failed to decode key: %w", err)
    }

    return nil
}
```

### Changes to `UnmarshalDynamoDBAttributeValue` method

```go
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
```

The key changes are:

1. In `UnmarshalMap`, instead of trying to unmarshal into a `row` struct that contains the interface, we extract the key fields and row data from the map, and then unmarshal the row data directly into the target struct.
2. In `UnmarshalDynamoDBAttributeValue`, we removed the code that was trying to unmarshal into the interface value, and instead just pass the key to the `DecodeKey` method.

These changes avoid the "unaddressable value" panic by not trying to modify the interface value directly.

## Additional Notes

This issue highlights the importance of understanding how Go's reflection system works with interfaces and pointers. When working with reflection-based libraries like the AWS SDK's attributevalue package, it's crucial to ensure that values are properly addressable when they need to be modified.

The fix ensures that:
1. The unmarshaling operation can properly modify the `row` struct.
2. The changes made during unmarshaling are properly propagated back to the original pointer.
3. The type information is preserved through the use of type assertions.

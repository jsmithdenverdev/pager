#!/bin/zsh

# Usage: ./export-dynamodb.sh <table-name> <output-file.json> [--region <region>]

TABLE_NAME=$1
OUTPUT_FILE=$2
REGION=${4:-us-west-2}

if [[ -z "$TABLE_NAME" || -z "$OUTPUT_FILE" ]]; then
  echo "Usage: $0 <table-name> <output-file.json> [--region <region>]"
  exit 1
fi

echo "Exporting data from DynamoDB table: $TABLE_NAME"

# Scan and transform
aws dynamodb scan \
  --table-name "$TABLE_NAME" \
  --region "$REGION" \
  --output json | \
jq '[.Items[] | { PutRequest: { Item: . } }]' > "$OUTPUT_FILE"

echo "✅ Export complete. File written to $OUTPUT_FILE"

package storage

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"strings"
	"time"
)

const tableName = "Resources"

type DynamoDBStore struct {
	svc *dynamodb.DynamoDB
}

func NewDynamoDBStore() *DynamoDBStore {
	// Create a new DynamoDBStore session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &DynamoDBStore{
		svc: dynamodb.New(sess),
	}
}

func (s *DynamoDBStore) GetItem(pk, sk string) (Resource, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {S: aws.String(pk)},
			"sk": {S: aws.String(sk)},
		},
	}

	result, err := s.svc.GetItem(input)
	if err != nil {
		var aErr awserr.Error
		if errors.As(err, &aErr) {
			if aErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return Resource{}, ErrNotFound
			}
		}

		return Resource{}, fmt.Errorf("failed to get item: %v", err)
	}

	var r Resource
	err = dynamodbattribute.UnmarshalMap(result.Item, &r)
	if err != nil {
		return Resource{}, fmt.Errorf("failed to unmarshal item: %v", err)
	}

	return r, nil
}

// QueryItems queries items from DynamoDB by sort key prefix
func (s *DynamoDBStore) QueryItems(pk, sk string, queryCondition QueryCondition) ([]Resource, error) {
	// Prepare the query input
	skPrefix := fmt.Sprintf("%v#", strings.Split(sk, "#")[0])
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String(string(queryCondition)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk":       {S: aws.String(pk)},
			":sk":       {S: aws.String(sk)},
			":skPrefix": {S: aws.String(skPrefix)},
		},
	}

	// Perform the query operation
	result, err := s.svc.Query(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}

	// Unmarshal the result into supplied argument
	var r []Resource
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal query result: %w", err)
	}

	return r, nil
}

// DeleteItem deletes an item from DynamoDB using the partition key and sort key
func (s *DynamoDBStore) DeleteItem(pk, sk string) error {
	// Prepare the input for the DeleteItem operation
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {S: aws.String(pk)},
			"sk": {S: aws.String(sk)},
		},
	}

	// Perform the delete operation
	_, err := s.svc.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// UpdateItem updates an item in DynamoDB using the provided partition key, sort key, and attributes to update.
func (s *DynamoDBStore) UpdateItem(pk, sk string, updates map[string]interface{}) error {
	// Build update expression
	var updateExprParts []string
	exprAttrValues := make(map[string]*dynamodb.AttributeValue)
	exprAttrNames := make(map[string]*string)

	for attr, value := range updates {
		nameToken := fmt.Sprintf("#%s", attr)
		updateExprParts = append(updateExprParts, fmt.Sprintf("%s = :%s", nameToken, attr))
		exprAttrNames[nameToken] = aws.String(attr)
		exprAttrValues[fmt.Sprintf(":%s", attr)] = &dynamodb.AttributeValue{
			S: aws.String(fmt.Sprintf("%v", value)), // Assumes string values; adjust as needed
		}
	}

	updateExpr := "SET " + strings.Join(updateExprParts, ", ")

	// Create the UpdateItemInput
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {S: aws.String(pk)},
			"sk": {S: aws.String(sk)},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames:  exprAttrNames,
		ReturnValues:              aws.String("UPDATED_NEW"),
	}

	// Perform the update
	_, err := s.svc.UpdateItem(input)
	if err != nil {
		var aErr awserr.Error
		if errors.As(err, &aErr) {
			if aErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return ErrNotFound
			}
		}

		return fmt.Errorf("failed to update item: %w", err)
	}

	return nil
}

// WriteItem Function to write an item to DynamoDBStore
func (s *DynamoDBStore) WriteItem(resource Resource) error {
	// Marshal the `value` argument into a map of DynamoDBStore attributes
	av, err := dynamodbattribute.MarshalMap(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal value to dynamodb attributes: %w", err)
	}

	// Add the partition key and sort key to the attribute map
	av["pk"] = &dynamodb.AttributeValue{S: aws.String(resource.PartitionKey)}
	av["sk"] = &dynamodb.AttributeValue{S: aws.String(resource.SortKey)}

	// Create the PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	// Put the item into DynamoDBStore
	_, err = s.svc.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to put item in dynamodb: %w", err)
	}

	return nil
}

func (s *DynamoDBStore) BatchWriteItems(resources []Resource) error {
	const maxBatchSize = 25

	// Split the items into batches of 25 (DynamoDB limit)
	for i := 0; i < len(resources); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(resources) {
			end = len(resources)
		}

		// Prepare the write requests for this batch
		var writeRequests []*dynamodb.WriteRequest
		for _, r := range resources[i:end] {
			av, err := dynamodbattribute.MarshalMap(r)
			if err != nil {
				return fmt.Errorf("failed to marshal r: %v", err)
			}
			av["pk"] = &dynamodb.AttributeValue{S: aws.String(r.PartitionKey)}
			av["sk"] = &dynamodb.AttributeValue{S: aws.String(r.SortKey)}

			writeRequests = append(writeRequests, &dynamodb.WriteRequest{
				PutRequest: &dynamodb.PutRequest{
					Item: av,
				},
			})
		}

		// Create the batch write input
		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				tableName: writeRequests,
			},
		}

		// Perform the batch write
		err := s.executeBatchWriteWithRetry(input)
		if err != nil {
			return fmt.Errorf("batch write failed: %v", err)
		}
	}

	return nil
}

// executeBatchWriteWithRetry executes a Write Item operation and retries if there are unprocessed items
func (s *DynamoDBStore) executeBatchWriteWithRetry(input *dynamodb.BatchWriteItemInput) error {
	for {
		// Perform the batch write operation
		result, err := s.svc.BatchWriteItem(input)
		if err != nil {
			return fmt.Errorf("failed to batch write items: %v", err)
		}

		// If there are no unprocessed items, return
		if len(result.UnprocessedItems) == 0 {
			return nil
		}

		// If there are unprocessed items, retry them after a delay
		input.RequestItems = result.UnprocessedItems
		time.Sleep(1 * time.Second) // Backoff before retrying
	}
}

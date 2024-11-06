package storage

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"time"
)

const tableName = "Resources"

type DynamoDBStorage struct {
	svc *dynamodb.DynamoDB
}

func NewDynamoDBStorage() *DynamoDBStorage {
	// Create a new DynamoDBStorage session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return &DynamoDBStorage{
		svc: dynamodb.New(sess),
	}
}

func (s *DynamoDBStorage) GetItem(pk, sk string, outPtr any) error {
	panicIfNotPointer(outPtr)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {S: aws.String(pk)},
			"sk": {S: aws.String(sk)},
		},
	}

	result, err := s.svc.GetItem(input)
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			if aerr.Code() == dynamodb.ErrCodeResourceNotFoundException {
				return ErrNotFound
			}
		}

		return fmt.Errorf("failed to get item: %v", err)
	}

	if result.Item == nil {
		return nil
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, outPtr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %v", err)
	}

	return nil
}

// QueryBySortKeyPrefix queries items from DynamoDB by sort key prefix
func (s *DynamoDBStorage) QueryBySortKeyPrefix(pk, skPrefix string, outSlicePtr any) error {
	panicIfNotSlicePointer(outSlicePtr)

	// Prepare the query input
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("pk = :pk AND begins_with(sk, :sk)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {S: aws.String(pk)},
			":sk": {S: aws.String(skPrefix)},
		},
	}

	// Perform the query operation
	result, err := s.svc.Query(input)
	if err != nil {
		return fmt.Errorf("failed to query items: %w", err)
	}

	// Unmarshal the result into supplied argument
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, outSlicePtr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal query result: %w", err)
	}

	return nil
}

// QueryItems queries items from DynamoDB by sort key prefix
func (s *DynamoDBStorage) QueryItems(pk, skPrefix string, queryCondition QueryCondition, outSlicePtr any) error {
	panicIfNotSlicePointer(outSlicePtr)
	panicIfInvalidQueryCondition(queryCondition)

	// Prepare the query input
	input := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String(string(queryCondition)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk":       {S: aws.String(pk)},
			":skPrefix": {S: aws.String(skPrefix)},
		},
	}

	// Perform the query operation
	result, err := s.svc.Query(input)
	if err != nil {
		return fmt.Errorf("failed to query items: %w", err)
	}

	// Unmarshal the result into supplied argument
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, outSlicePtr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal query result: %w", err)
	}

	return nil
}

// DeleteItem deletes an item from DynamoDB using the partition key and sort key
func (s *DynamoDBStorage) DeleteItem(pk, sk string) error {
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

// WriteItem Function to write an item to DynamoDBStorage
func (s *DynamoDBStorage) WriteItem(item TableItem) error {
	// Marshal the `value` argument into a map of DynamoDBStorage attributes
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal value to dynamodb attributes: %w", err)
	}

	// Add the partition key and sort key to the attribute map
	av["pk"] = &dynamodb.AttributeValue{S: aws.String(item.GetPartitionKey())}
	av["sk"] = &dynamodb.AttributeValue{S: aws.String(item.GetSortKey())}

	// Create the PutItem input
	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	// Put the item into DynamoDBStorage
	_, err = s.svc.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to put item in dynamodb: %w", err)
	}

	return nil
}

func (s *DynamoDBStorage) BatchWriteItems(items []TableItem) error {
	const maxBatchSize = 25

	// Split the items into batches of 25 (DynamoDB limit)
	for i := 0; i < len(items); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(items) {
			end = len(items)
		}

		// Prepare the write requests for this batch
		var writeRequests []*dynamodb.WriteRequest
		for _, item := range items[i:end] {
			av, err := dynamodbattribute.MarshalMap(item)
			if err != nil {
				return fmt.Errorf("failed to marshal item: %v", err)
			}
			av["pk"] = &dynamodb.AttributeValue{S: aws.String(item.GetPartitionKey())}
			av["sk"] = &dynamodb.AttributeValue{S: aws.String(item.GetSortKey())}

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
		err := executeBatchWriteWithRetry(s.svc, input)
		if err != nil {
			return fmt.Errorf("batch write failed: %v", err)
		}
	}

	return nil
}

// executeBatchWriteWithRetry executes a Item operation and retries if there are unprocessed items
func executeBatchWriteWithRetry(svc *dynamodb.DynamoDB, input *dynamodb.BatchWriteItemInput) error {
	for {
		// Perform the batch write operation
		result, err := svc.BatchWriteItem(input)
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

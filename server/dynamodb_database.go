package server

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBDatabase struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBDatabase(client *dynamodb.Client, tableName string) (*DynamoDBDatabase, error) {
	return &DynamoDBDatabase{
		client:    client,
		tableName: tableName,
	}, nil
}

func dynamoDBActivityHashKey(locale *Locale) []byte {
	return []byte("activity_by_locale:" + locale.Subdomain)
}

func (db *DynamoDBDatabase) AddActivity(activity []Activity) error {
	for _, locale := range Locales {
		var remaining []Activity
		for _, a := range activity {
			if locale.ActivityFilter(a) {
				remaining = append(remaining, a)
			}
		}

		for len(remaining) > 0 {
			batch := remaining
			const maxBatchSize = 25
			if len(batch) > maxBatchSize {
				batch = batch[:maxBatchSize]
			}

			writeRequests := make([]dynamodb.WriteRequest, len(batch))
			for i, a := range batch {
				k, v, err := marshalActivity(a)
				if err != nil {
					return err
				}
				writeRequests[i] = dynamodb.WriteRequest{
					PutRequest: &dynamodb.PutRequest{
						Item: map[string]dynamodb.AttributeValue{
							"hk": dynamodb.AttributeValue{
								B: dynamoDBActivityHashKey(locale),
							},
							"rk": dynamodb.AttributeValue{
								B: []byte(k),
							},
							"v": dynamodb.AttributeValue{
								B: []byte(v),
							},
						},
					},
				}
			}
			unprocessed := map[string][]dynamodb.WriteRequest{
				db.tableName: writeRequests,
			}

			for len(unprocessed) > 0 {
				result, err := db.client.BatchWriteItemRequest(&dynamodb.BatchWriteItemInput{
					RequestItems: unprocessed,
				}).Send(context.Background())
				if err != nil {
					return err
				}
				unprocessed = result.UnprocessedItems
			}

			remaining = remaining[len(batch):]
		}
	}
	return nil
}

func (db *DynamoDBDatabase) Activity(locale *Locale, start string, count int) ([]Activity, string, error) {
	var activity []Activity

	var startKey map[string]dynamodb.AttributeValue
	if start != "" {
		rk, _ := base64.RawURLEncoding.DecodeString(start)
		startKey = map[string]dynamodb.AttributeValue{
			"hk": dynamodb.AttributeValue{
				B: dynamoDBActivityHashKey(locale),
			},
			"rk": dynamodb.AttributeValue{
				B: rk,
			},
		}
	}

	condition := "hk = :hash"
	attributeValues := map[string]dynamodb.AttributeValue{
		":hash": dynamodb.AttributeValue{
			B: dynamoDBActivityHashKey(locale),
		},
	}

	for len(activity) < count {
		result, err := db.client.QueryRequest(&dynamodb.QueryInput{
			TableName:                 aws.String(db.tableName),
			KeyConditionExpression:    aws.String(condition),
			ExpressionAttributeValues: attributeValues,
			ExclusiveStartKey:         startKey,
			Limit:                     aws.Int64(int64(count - len(activity))),
			ScanIndexForward:          aws.Bool(false),
		}).Send(context.Background())
		if err != nil {
			return nil, "", err
		}
		for _, item := range result.Items {
			if a, err := unmarshalActivity(item["rk"].B, item["v"].B); err != nil {
				return nil, "", err
			} else if a != nil {
				activity = append(activity, a)
			}
		}
		if result.LastEvaluatedKey == nil {
			break
		}
		startKey = result.LastEvaluatedKey
	}

	var next string
	if startKey != nil {
		next = base64.RawURLEncoding.EncodeToString(startKey["rk"].B)
	}
	return activity, next, nil
}

func (db *DynamoDBDatabase) Close() error {
	return nil
}

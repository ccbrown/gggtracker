package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/defaults"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/require"
)

func newDynamoDBTestClient() (*dynamodb.Client, error) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")

	config := defaults.Config()
	config.Region = "us-east-1"
	config.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if endpoint != "" {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		}
		return aws.Endpoint{
			URL: "http://localhost:8000",
		}, nil
	})

	credentialsBuf := make([]byte, 20)
	if _, err := rand.Read(credentialsBuf); err != nil {
		return nil, err
	}
	credentials := base64.RawURLEncoding.EncodeToString(credentialsBuf)
	config.Credentials = aws.NewStaticCredentialsProvider(credentials, credentials, "")
	config.Retryer = aws.DefaultRetryer{
		NumMaxRetries: 0,
	}

	client := dynamodb.New(config)
	if endpoint == "" {
		if _, err := client.ListTablesRequest(&dynamodb.ListTablesInput{}).Send(context.Background()); err != nil {
			if err, ok := err.(awserr.Error); ok && err.Code() == "RequestError" {
				return nil, nil
			}
		}
	}
	return client, nil
}

func createDynamoDBTable(client *dynamodb.Client, tableName string) error {
	if _, err := client.CreateTableRequest(&dynamodb.CreateTableInput{
		AttributeDefinitions: []dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("hk"),
				AttributeType: dynamodb.ScalarAttributeTypeB,
			}, {
				AttributeName: aws.String("rk"),
				AttributeType: dynamodb.ScalarAttributeTypeB,
			},
		},
		KeySchema: []dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("hk"),
				KeyType:       dynamodb.KeyTypeHash,
			}, {
				AttributeName: aws.String("rk"),
				KeyType:       dynamodb.KeyTypeRange,
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(25),
			WriteCapacityUnits: aws.Int64(25),
		},
		TableName: &tableName,
	}).Send(context.Background()); err != nil {
		return err
	}
	return client.WaitUntilTableExists(context.Background(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
}

func TestDynamoDBDatabase(t *testing.T) {
	client, err := newDynamoDBTestClient()
	require.NoError(t, err)
	if client == nil {
		t.Skip("launch a local dynamodb container to run this test: docker run --rm -it -p 8000:8000 dwmkerr/dynamodb -inMemory")
	}

	const tableName = "TestDynamoDBDatabase"

	if _, err := client.DeleteTableRequest(&dynamodb.DeleteTableInput{
		TableName: aws.String(tableName),
	}).Send(context.Background()); err == nil {
		require.NoError(t, client.WaitUntilTableNotExists(context.Background(), &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}))
	}

	require.NoError(t, createDynamoDBTable(client, tableName))

	db, err := NewDynamoDBDatabase(client, tableName)
	require.NoError(t, err)
	defer db.Close()

	testDatabase(t, db)
}

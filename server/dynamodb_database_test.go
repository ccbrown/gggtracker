package server

import (
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

func newDynamoDBTestClient() (*dynamodb.DynamoDB, error) {
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
		if _, err := client.ListTablesRequest(&dynamodb.ListTablesInput{}).Send(); err != nil {
			if err, ok := err.(awserr.Error); ok && err.Code() == "RequestError" {
				return nil, nil
			}
		}
	}
	return client, nil
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
	}).Send(); err == nil {
		require.NoError(t, client.WaitUntilTableNotExists(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		}))
	}

	require.NoError(t, CreateDynamoDBTable(client, tableName))

	db, err := NewDynamoDBDatabase(client, tableName)
	require.NoError(t, err)
	defer db.Close()

	testDatabase(t, db)
}

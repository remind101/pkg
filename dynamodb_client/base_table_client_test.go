package dynamodb_client_test

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	. "github.com/remind101/pkg/dynamodb_client"
	"github.com/stretchr/testify/assert"
)

var testTableName = "test_table"
var AllTableDescriptions = map[string]*dynamodb.TableDescription{
	testTableName: &dynamodb.TableDescription{
		TableName: aws.String(testTableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String("user_uuid"),
				AttributeType: aws.String("S"),
			},
			&dynamodb.AttributeDefinition{
				AttributeName: aws.String("range_uuid"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String("user_uuid"),
				KeyType:       aws.String("HASH"),
			},
			&dynamodb.KeySchemaElement{
				AttributeName: aws.String("range_uuid"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughputDescription{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
	},
}

func fetchEnv(key string, dval string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return dval
}

func newFreshDynamoClient() *DynamoClient {
	params := DynamoConnectionParams{
		RegionName:        "us-west-2",
		LocalDynamoURL:    fetchEnv("LOCAL_DYNAMO_URL", "http://127.0.0.1:8000"),
		Scope:             "client-test",
		TableDescriptions: AllTableDescriptions,
	}
	c := NewDynamoDBClient(session.New(), params)
	c.DeleteTables()
	c.CreateTables()
	return c
}

func TestTableClient(t *testing.T) {
	dClient := newFreshDynamoClient()
	_, err := dClient.GetClientForTable("client-test")
	assert.NotNil(t, err)
	_, err = dClient.GetClientForTable(testTableName)
	assert.Nil(t, err)
}

func TestCRUDItem(t *testing.T) {
	dClient := newFreshDynamoClient()
	tClient, _ := dClient.GetClientForTable(testTableName)
	data, err := tClient.GetItem(context.Background(), "user-a", "range-a")
	assert.Nil(t, data)
	data, err = dynamodbattribute.MarshalMap(map[string]string{
		"user_uuid":  "user-a",
		"range_uuid": "range-a",
		"beans":      "cool",
	})
	assert.Nil(t, err)
	err = tClient.PutItem(context.Background(), "user-a", "range-a", data)
	assert.Nil(t, err)
	data, err = tClient.GetItem(context.Background(), "user-a", "range-a")
	assert.NotNil(t, data)
	var parsedData map[string]string
	err = dynamodbattribute.UnmarshalMap(data, &parsedData)
	assert.Nil(t, err)
	assert.Equal(t, parsedData["beans"], "cool")
}

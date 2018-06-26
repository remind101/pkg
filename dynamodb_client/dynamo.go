package dynamodb_client

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/remind101/pkg/logger"
)

type DynamoClient struct {
	dynamodbiface.DynamoDBAPI
	TableDescriptions map[string]*dynamodb.TableDescription
	ServiceName       string
}

type DynamoConnectionParams struct {
	RegionName        string
	LocalDynamoURL    string // if not "", URL of local Dynamo to point to.
	Scope             string // configures the table name per env
	TableDescriptions map[string]*dynamodb.TableDescription
	ServiceName       string // Name to use in apm
}

func NewDynamoDBClient(c client.ConfigProvider, params DynamoConnectionParams) *DynamoClient {
	if params.RegionName == "" {
		params.RegionName = "us-east-1"
	}
	config := aws.Config{
		Region:   aws.String(params.RegionName),
		Endpoint: aws.String(params.LocalDynamoURL),
	}
	svc := dynamodb.New(c, &config)
	//svc.Handlers.Retry.PushFrontNamed(CheckThrottleHandler)

	return &DynamoClient{
		svc,
		scopedTabledDescriptions(params.TableDescriptions, params.Scope),
		params.ServiceName,
	}
}

func scopedTabledDescriptions(tds map[string]*dynamodb.TableDescription, scope string) map[string]*dynamodb.TableDescription {
	scopedTds := make(map[string]*dynamodb.TableDescription, len(tds))
	for name, td := range tds {
		newTd := *td
		newTd.TableName = aws.String(fmt.Sprintf("%s-%s", scope, name))
		scopedTds[name] = &newTd
	}
	return scopedTds
}

type DynamoKey struct {
	HashKey  string
	RangeKey string
}

type AwsExpression struct {
	// http://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#UpdateItemInput
	ConditionExpression       string
	UpdateExpression          string
	ExpressionAttributeNames  map[string]*string
	ExpressionAttributeValues map[string]*dynamodb.AttributeValue
}

type QueryOpts struct {
	HashKey                   string
	Limit                     int64
	Descending                bool
	ExpressionAttributeNames  map[string]*string
	ExpressionAttributeValues map[string]*dynamodb.AttributeValue
	ProjectionExpression      string
	KeyConditionExpression    string
	FilterExpression          string
	IndexName                 string
}

func (dc *DynamoClient) GetClientForTable(name string) (TableClient, error) {
	if val, ok := dc.TableDescriptions[name]; ok {
		return NewBaseTableClient(TableClientOpts{
			Desc:        val,
			Client:      dc,
			ServiceName: dc.ServiceName,
		})
	} else {
		return nil, fmt.Errorf("Table not found")
	}
}

func (dc *DynamoClient) CreateTables() error {
	for tableName, _ := range dc.TableDescriptions {
		tableClient, err := dc.GetClientForTable(tableName)
		if err != nil {
			return err
		}
		fmt.Printf("Creating table %s \n", tableName)
		err = tableClient.Create(context.Background())
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				fmt.Printf("%s and %s\n", awsErr.Error(), awsErr.Code())
				if awsErr.Code() == "ResourceInUseException" { // table already exists
					logger.Debug(context.Background(), "table", tableName, "already exists")
					continue
				}
			}
			return fmt.Errorf("Error creating table %s: %v", tableName, err)
		}
		logger.Debug(context.Background(), "table", tableName, "created successfully")
	}
	return nil
}

func (dc *DynamoClient) DeleteTables() error {
	for tableName, _ := range dc.TableDescriptions {
		tableClient, err := dc.GetClientForTable(tableName)
		if err != nil {
			return err
		}
		fmt.Printf("Deleting table %s \n", tableName)
		tableClient.Delete(context.Background())
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				fmt.Printf("%s and %s\n", awsErr.Error(), awsErr.Code())
				if awsErr.Code() == "ResourceInUseException" { // table doesn't exist
					logger.Debug(context.Background(), "table", tableName, "doesn't exist")
					continue
				}
			}
			return fmt.Errorf("Error creating table %s: %v", tableName, err)
		}
		logger.Debug(context.Background(), "table", tableName, "created successfully")
	}
	return nil
}

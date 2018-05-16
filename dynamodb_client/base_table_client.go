package dynamodb_client

import (
	"context"
	"fmt"
	"strconv"

	dd_opentracing "github.com/DataDog/dd-trace-go/opentracing"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	opentracing "github.com/opentracing/opentracing-go"
)

var (
	BuildQueryHashKeyName = ":hash_key"
	AwsErrNotFound        = fmt.Errorf("item not found")
	AwsErrNotProcessed    = fmt.Errorf("item not processed")
)

type PrimaryKey struct {
	PartitionKey *Key
	SortKey      *Key
}

type Key struct {
	Type string
	Name string
}

// BaseTableClient implements the TableClient interface.
type BaseTableClient struct {
	desc        *dynamodb.TableDescription
	client      dynamodbiface.DynamoDBAPI
	pk          *PrimaryKey
	tableName   string
	serviceName string
}

// TableClient defines a context aware table based interface to manage a dynamodb
// table.
type TableClient interface {
	// Reads
	GetItem(ctx context.Context, hashKey, rangeKey string) (DynamoItem, error)
	// Writes
	PutItem(ctx context.Context, hashKey, rangeKey string, item DynamoItem) error
	DeleteItem(ctx context.Context, hashKey, rangeKey string) error
	// Table management
	Create(ctx context.Context) error
	Delete(ctx context.Context) error

	// LOOKING FOR SOMETHING???
	// This implementation was pulled from another project. The functions above have had tests added.
	// If you want to use any of the functionality below, please uncomment the declaration and write a test
	// for it. The implementation is already in this file, though without tests we don't know if it is correct.

	// Reads
	// Query(ctx context.Context, opts QueryOpts) ([]map[string]*dynamodb.AttributeValue, error)
	//	BatchGetDocument(ctx context.Context, keys []DynamoKey, consistentRead bool, v interface{}) ([]error, error)
	// Writes
	//	ConditionExpressionPutItem(ctx context.Context, hashKey, rangeKey string, item DynamoItem, expression AwsExpression) error
	//	ConditionExpressionDeleteItem(ctx context.Context, hashKey, rangeKey string, expr AwsExpression) error
	//	ConditionExpressionUpdateAttributes(ctx context.Context, hashKey, rangeKey string, expr AwsExpression) error
	//	BatchPutDocument(ctx context.Context, keys []DynamoKey, items []DynamoItem) ([]error, error)
	//	BatchDeleteDocument(ctx context.Context, keys []DynamoKey) ([]error, error)
	// Table Management
	//	UpdateStreamConfiguration(ctx context.Context) error
}

type TableClientOpts struct {
	Desc        *dynamodb.TableDescription
	Client      dynamodbiface.DynamoDBAPI
	ServiceName string
}

func NewBaseTableClient(opts TableClientOpts) (*BaseTableClient, error) {
	tableName := aws.StringValue(opts.Desc.TableName)
	pk, err := buildPrimaryKey(opts.Desc)
	if err != nil {
		return nil, fmt.Errorf("Error building primary key for %s: %v", tableName, err)
	}

	table := &BaseTableClient{
		desc:        opts.Desc,
		client:      opts.Client,
		pk:          &pk,
		tableName:   tableName,
		serviceName: opts.ServiceName,
	}

	return table, nil
}

func (dt *BaseTableClient) Query(ctx context.Context, opts QueryOpts) ([]map[string]*dynamodb.AttributeValue, error) {
	input := dt.buildQuery(opts)

	req, output := dt.client.QueryRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return output.Items, err
}

func (dt *BaseTableClient) PutItem(ctx context.Context, hashKey, rangeKey string, item DynamoItem) error {
	item = dt.addPrimaryKey(hashKey, rangeKey, item)

	input := &dynamodb.PutItemInput{
		Item:      item,
		TableName: dt.desc.TableName,
	}

	req, _ := dt.client.PutItemRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return err
}

func (dt *BaseTableClient) DeleteItem(ctx context.Context, hashKey, rangeKey string) error {
	key := dt.addPrimaryKey(hashKey, rangeKey, DynamoItem{})

	input := &dynamodb.DeleteItemInput{
		TableName: dt.desc.TableName,
		Key:       key,
	}

	req, _ := dt.client.DeleteItemRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return err
}

//func (dt *BaseTableClient) ConditionExpressionDeleteItem(ctx context.Context, hashKey, rangeKey string, expr AwsExpression) error {
//	key := dt.addPrimaryKey(hashKey, rangeKey, DynamoItem{})
//
//	input := &dynamodb.DeleteItemInput{
//		TableName:                 dt.desc.TableName,
//		Key:                       key,
//		ConditionExpression:       aws.String(expr.ConditionExpression),
//		ExpressionAttributeNames:  expr.ExpressionAttributeNames,
//		ExpressionAttributeValues: expr.ExpressionAttributeValues,
//	}
//
//	req, output := dt.client.DeleteItemRequest(input)
//	err := dt.sendWithTracing(ctx, req)
//
//	return err
//}
//
//func (dt *BaseTableClient) ConditionExpressionPutItem(ctx context.Context, hashKey, rangeKey string, item DynamoItem, expr AwsExpression) error {
//	item = dt.addPrimaryKey(hashKey, rangeKey, item)
//
//	input := &dynamodb.PutItemInput{
//		TableName:                 dt.desc.TableName,
//		Item:                      item,
//		ConditionExpression:       aws.String(expr.ConditionExpression),
//		ExpressionAttributeNames:  expr.ExpressionAttributeNames,
//		ExpressionAttributeValues: expr.ExpressionAttributeValues,
//	}
//
//	req, output := dt.client.PutItemRequest(input)
//	err := dt.sendWithTracing(ctx, req)
//
//	return err
//}
//
//func (dt *BaseTableClient) ConditionExpressionUpdateAttributes(ctx context.Context, hashKey, rangeKey string, expr AwsExpression) error {
//	key := dt.addPrimaryKey(hashKey, rangeKey, DynamoItem{})
//
//	input := &dynamodb.UpdateItemInput{
//		TableName:                 dt.desc.TableName,
//		Key:                       key,
//		ConditionExpression:       aws.String(expr.ConditionExpression),
//		ExpressionAttributeNames:  expr.ExpressionAttributeNames,
//		ExpressionAttributeValues: expr.ExpressionAttributeValues,
//		UpdateExpression:          aws.String(expr.UpdateExpression),
//	}
//
//	req, output := dt.client.UpdateItemRequest(input)
//	err := dt.sendWithTracing(ctx, req)
//
//	return err
//}
//
func (dt *BaseTableClient) GetItem(ctx context.Context, hashKey, rangeKey string) (DynamoItem, error) {
	item := dt.addPrimaryKey(hashKey, rangeKey, DynamoItem{})

	input := &dynamodb.GetItemInput{
		TableName:      dt.desc.TableName,
		Key:            item,
		ConsistentRead: aws.Bool(false),
	}

	req, output := dt.client.GetItemRequest(input)
	err := dt.sendWithTracing(ctx, req)

	if output != nil {
		// Do we still want this?
		//dt.recordConsumedCapacity(ctx, "GetItem", hashKey, output.ConsumedCapacity)
	}

	if isEmptyGetItemOutput(output) {
		return nil, AwsErrNotFound
	}

	return output.Item, err
}

func (dt *BaseTableClient) Create(ctx context.Context) error {
	pt := &dynamodb.ProvisionedThroughput{
		ReadCapacityUnits:  dt.desc.ProvisionedThroughput.ReadCapacityUnits,
		WriteCapacityUnits: dt.desc.ProvisionedThroughput.WriteCapacityUnits,
	}

	var localSecondaryIndexes []*dynamodb.LocalSecondaryIndex
	for _, desc := range dt.desc.LocalSecondaryIndexes {
		localSecondaryIndexes = append(localSecondaryIndexes, &dynamodb.LocalSecondaryIndex{
			IndexName:  desc.IndexName,
			KeySchema:  desc.KeySchema,
			Projection: desc.Projection,
		})
	}

	var globalSecondaryIndexes []*dynamodb.GlobalSecondaryIndex
	for _, desc := range dt.desc.GlobalSecondaryIndexes {
		gpt := &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  desc.ProvisionedThroughput.ReadCapacityUnits,
			WriteCapacityUnits: desc.ProvisionedThroughput.WriteCapacityUnits,
		}
		globalSecondaryIndexes = append(globalSecondaryIndexes, &dynamodb.GlobalSecondaryIndex{
			IndexName:             desc.IndexName,
			KeySchema:             desc.KeySchema,
			Projection:            desc.Projection,
			ProvisionedThroughput: gpt,
		})
	}

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions:   dt.desc.AttributeDefinitions,
		ProvisionedThroughput:  pt,
		KeySchema:              dt.desc.KeySchema,
		TableName:              dt.desc.TableName,
		LocalSecondaryIndexes:  localSecondaryIndexes,
		GlobalSecondaryIndexes: globalSecondaryIndexes,
		StreamSpecification:    dt.desc.StreamSpecification,
	}

	req, _ := dt.client.CreateTableRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return err
}

func (dt *BaseTableClient) Delete(ctx context.Context) error {
	input := &dynamodb.DeleteTableInput{
		TableName: dt.desc.TableName,
	}

	req, _ := dt.client.DeleteTableRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return err
}

//func (dt *BaseTableClient) UpdateStreamConfiguration(ctx context.Context) error {
//	if dt.desc.StreamSpecification == nil {
//		return nil
//	}
//	tableName := *dt.desc.TableName
//	enable := dt.desc.StreamSpecification.StreamEnabled != nil && *dt.desc.StreamSpecification.StreamEnabled
//	input := &dynamodb.UpdateTableInput{
//		TableName:           dt.desc.TableName,
//		StreamSpecification: dt.desc.StreamSpecification,
//	}
//
//	req, _ := dt.client.UpdateTableRequest(input)
//	err := dt.sendWithTracing(ctx, req)
//
//	if err != nil {
//		if awsErr, ok := err.(awserr.Error); ok {
//			if awsErr.Code() == "ValidationException" { // stream already configured
//				if enable {
//					logger.Debug(ctx, "stream", tableName, "already enabled")
//				} else {
//					logger.Debug(ctx, "stream", tableName, "already disabled")
//				}
//				return nil
//			}
//		}
//		return fmt.Errorf("Error creating stream %s: %v", tableName, err)
//	}
//
//	if enable {
//		logger.Debug(ctx, "stream", tableName, "enabled")
//	} else {
//		logger.Debug(ctx, "stream", tableName, "disabled")
//	}
//	return nil
//}

func (dt *BaseTableClient) BatchGetDocument(ctx context.Context, keys []DynamoKey, consistentRead bool) (*dynamodb.BatchGetItemOutput, *request.Request, error) {
	keysSlice := make([]map[string]*dynamodb.AttributeValue, len(keys))
	for i, key := range keys {
		keysSlice[i] = dt.addPrimaryKey(key.HashKey, key.RangeKey, DynamoItem{})
	}

	requestItems := map[string]*dynamodb.KeysAndAttributes{
		*dt.desc.TableName: &dynamodb.KeysAndAttributes{
			ConsistentRead: aws.Bool(consistentRead),
			Keys:           keysSlice,
		},
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: requestItems,
	}

	req, output := dt.client.BatchGetItemRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return output, req, err
}

func (dt *BaseTableClient) BatchPutDocument(ctx context.Context, keys []DynamoKey, items []DynamoItem) (*dynamodb.BatchWriteItemOutput, *request.Request, error) {
	if len(keys) != len(items) {
		return nil, nil, fmt.Errorf("keys and items must have same length")
	}

	writeRequests := make([]*dynamodb.WriteRequest, len(keys))
	for index, key := range keys {
		item := dt.addPrimaryKey(key.HashKey, key.RangeKey, items[index])

		writeRequests[index] = &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: item,
			},
		}
	}

	requestItems := map[string][]*dynamodb.WriteRequest{
		*dt.desc.TableName: writeRequests,
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: requestItems,
	}

	req, output := dt.client.BatchWriteItemRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return output, req, err
}

func (dt *BaseTableClient) BatchDeleteDocument(ctx context.Context, keys []DynamoKey) (*dynamodb.BatchWriteItemOutput, *request.Request, error) {
	writeRequests := make([]*dynamodb.WriteRequest, len(keys))
	for i, key := range keys {
		item := dt.addPrimaryKey(key.HashKey, key.RangeKey, DynamoItem{})

		writeRequests[i] = &dynamodb.WriteRequest{
			DeleteRequest: &dynamodb.DeleteRequest{
				Key: item,
			},
		}
	}

	requestItems := map[string][]*dynamodb.WriteRequest{
		*dt.desc.TableName: writeRequests,
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: requestItems,
	}

	req, output := dt.client.BatchWriteItemRequest(input)
	err := dt.sendWithTracing(ctx, req)

	return output, req, err
}

// Builds a PrimaryKey from a TableDescription
func buildPrimaryKey(t *dynamodb.TableDescription) (pk PrimaryKey, err error) {
	for _, k := range t.KeySchema {
		ad := findAttributeDefinitionByName(t.AttributeDefinitions, aws.StringValue(k.AttributeName))
		if ad == nil {
			return pk, fmt.Errorf("An inconsistency found in TableDescription")
		}

		switch aws.StringValue(k.KeyType) {
		case dynamodb.KeyTypeHash:
			pk.PartitionKey = &Key{Type: aws.StringValue(ad.AttributeType), Name: aws.StringValue(k.AttributeName)}
		case dynamodb.KeyTypeRange:
			pk.SortKey = &Key{Type: aws.StringValue(ad.AttributeType), Name: aws.StringValue(k.AttributeName)}
		default:
			return pk, fmt.Errorf("key type not supported")
		}
	}
	return
}

// Finds Attribute Definition matching the passed in name
func findAttributeDefinitionByName(ads []*dynamodb.AttributeDefinition, name string) *dynamodb.AttributeDefinition {
	for _, ad := range ads {
		if aws.StringValue(ad.AttributeName) == name {
			return ad
		}
	}
	return nil
}

// Adds HashKey and RangeKey to a DynamoItem
func (dt *BaseTableClient) addPrimaryKey(hashKey, rangeKey string, item DynamoItem) DynamoItem {
	item[dt.pk.PartitionKey.Name] = dt.pk.PartitionKey.newAttributeValue(hashKey)

	if dt.pk.hasSortKey() {
		item[dt.pk.SortKey.Name] = dt.pk.SortKey.newAttributeValue(rangeKey)
	}

	return item
}

func (pk *PrimaryKey) hasSortKey() bool {
	if pk.SortKey != nil {
		return true
	}
	return false
}

// Builds an attribute value from a PrimaryKey attribute
func (k *Key) newAttributeValue(value string) *dynamodb.AttributeValue {
	switch k.Type {
	case dynamodb.ScalarAttributeTypeS:
		return NewStringAttributeValue(value)
	case dynamodb.ScalarAttributeTypeN:
		return NewNumberAttributeValue(value)
	case dynamodb.ScalarAttributeTypeB:
		b, _ := strconv.ParseBool(value)
		return NewBoolAttributeValue(b)
	default:
		return nil
	}
}

func isEmptyGetItemOutput(gio *dynamodb.GetItemOutput) bool {
	return gio.Item == nil
}

func isValidConsumedCapacityLevel(level string) bool {
	switch level {
	case dynamodb.ReturnConsumedCapacityIndexes:
		return true
	case dynamodb.ReturnConsumedCapacityTotal:
		return true
	case dynamodb.ReturnConsumedCapacityNone:
		return true
	default:
		return false
	}
}

func (dt *BaseTableClient) buildQuery(opts QueryOpts) *dynamodb.QueryInput {
	qi := &dynamodb.QueryInput{TableName: dt.desc.TableName}

	// Copy over most fields from opts
	if opts.Limit != 0 {
		qi.Limit = aws.Int64(opts.Limit)
	}
	if opts.Descending {
		qi.ScanIndexForward = aws.Bool(false)
	}
	if opts.IndexName != "" {
		qi.IndexName = aws.String(opts.IndexName)
	}

	if opts.FilterExpression != "" {
		qi.FilterExpression = aws.String(opts.FilterExpression)
	}

	if opts.ProjectionExpression != "" {
		qi.ProjectionExpression = aws.String(opts.ProjectionExpression)
	}
	qi.ExpressionAttributeValues = opts.ExpressionAttributeValues
	qi.ExpressionAttributeNames = opts.ExpressionAttributeNames

	// HashKeys are added as key conditions
	keyCondition := ""
	if opts.HashKey != "" {
		keyCondition += fmt.Sprintf("%s = %s", dt.pk.PartitionKey.Name, BuildQueryHashKeyName)

		if qi.ExpressionAttributeValues == nil {
			qi.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{}
		}
		qi.ExpressionAttributeValues[BuildQueryHashKeyName] = NewStringAttributeValue(opts.HashKey)

		if opts.KeyConditionExpression != "" {
			keyCondition += " AND "
		}
	}
	keyCondition += opts.KeyConditionExpression
	qi.KeyConditionExpression = aws.String(keyCondition)
	return qi
}

func (dt *BaseTableClient) sendWithTracing(ctx context.Context, r *request.Request) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "client.request")
	defer span.Finish()
	r.HTTPRequest = r.HTTPRequest.WithContext(ctx)

	span.SetTag(dd_opentracing.SpanType, "db")
	span.SetTag(dd_opentracing.ServiceName, dt.serviceName)
	span.SetTag(dd_opentracing.ResourceName, r.Operation.Name)
	span.SetTag("http.method", r.Operation.HTTPMethod)
	span.SetTag("http.url", r.ClientInfo.Endpoint+r.Operation.HTTPPath)
	span.SetTag("out.host", r.ClientInfo.Endpoint)
	span.SetTag("aws.operation", r.Operation.Name)
	span.SetTag("aws.table_name", dt.tableName)

	err := r.Send()

	span.SetTag("aws.retry_count", r.RetryCount)

	if r.HTTPResponse != nil {
		span.SetTag("http.status_code", r.HTTPResponse.StatusCode)
	}

	if err != nil {
		span.SetTag(dd_opentracing.Error, err)
	}

	return err
}

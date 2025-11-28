package db

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDB struct {
	Client        *dynamodb.Client
	UsersTable    string
	RoomsTable    string
	MessagesTable string
	CodeSyncTable string
}

func NewDynamoDB() (*DynamoDB, error) {
	region := os.Getenv("AWS_REGION")
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	db := &DynamoDB{
		Client:        client,
		UsersTable:    os.Getenv("DYNAMO_USERS_TABLE"),
		RoomsTable:    os.Getenv("DYNAMO_ROOMS_TABLE"),
		MessagesTable: os.Getenv("DYNAMO_MESSAGES_TABLE"),
		CodeSyncTable: os.Getenv("DYNAMO_CODESYNC_TABLE"),
	}

	log.Printf("DynamoDB client initialized (Region: %s)", region)
	return db, nil
}

func (db *DynamoDB) EnsureTablesExist(ctx context.Context) error {
	tables := []struct {
		Name string
		Key  []types.KeySchemaElement
		Attr []types.AttributeDefinition
		GSI  []types.GlobalSecondaryIndex
	}{
		{
			Name: db.UsersTable,
			Key: []types.KeySchemaElement{
				{AttributeName: aws.String("userId"), KeyType: types.KeyTypeHash},
			},
			Attr: []types.AttributeDefinition{
				{AttributeName: aws.String("userId"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("email"), AttributeType: types.ScalarAttributeTypeS},
			},
			GSI: []types.GlobalSecondaryIndex{
				{
					IndexName: aws.String("EmailIndex"),
					KeySchema: []types.KeySchemaElement{
						{AttributeName: aws.String("email"), KeyType: types.KeyTypeHash},
					},
					Projection: &types.Projection{
						ProjectionType: types.ProjectionTypeAll,
					},
					ProvisionedThroughput: &types.ProvisionedThroughput{
						ReadCapacityUnits:  aws.Int64(1),
						WriteCapacityUnits: aws.Int64(1),
					},
				},
			},
		},
		{
			Name: db.RoomsTable,
			Key: []types.KeySchemaElement{
				{AttributeName: aws.String("roomId"), KeyType: types.KeyTypeHash},
			},
			Attr: []types.AttributeDefinition{
				{AttributeName: aws.String("roomId"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
		{
			Name: db.MessagesTable,
			Key: []types.KeySchemaElement{
				{AttributeName: aws.String("roomId"), KeyType: types.KeyTypeHash},
				{AttributeName: aws.String("timestamp"), KeyType: types.KeyTypeRange},
			},
			Attr: []types.AttributeDefinition{
				{AttributeName: aws.String("roomId"), AttributeType: types.ScalarAttributeTypeS},
				{AttributeName: aws.String("timestamp"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
		{
			Name: db.CodeSyncTable,
			Key: []types.KeySchemaElement{
				{AttributeName: aws.String("roomId"), KeyType: types.KeyTypeHash},
			},
			Attr: []types.AttributeDefinition{
				{AttributeName: aws.String("roomId"), AttributeType: types.ScalarAttributeTypeS},
			},
		},
	}

	listTables, err := db.Client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		return err
	}

	existingTables := make(map[string]bool)
	for _, name := range listTables.TableNames {
		existingTables[name] = true
	}

	for _, table := range tables {
		if existingTables[table.Name] {
			log.Printf("Table %s already exists", table.Name)
			continue
		}

		log.Printf("📦 Creating table %s...", table.Name)
		_, err := db.Client.CreateTable(ctx, &dynamodb.CreateTableInput{
			TableName:              aws.String(table.Name),
			KeySchema:              table.Key,
			AttributeDefinitions:   table.Attr,
			GlobalSecondaryIndexes: table.GSI,
			ProvisionedThroughput: &types.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
		})
		if err != nil {
			return err
		}
		log.Printf("Table %s created successfully", table.Name)
	}

	return nil
}

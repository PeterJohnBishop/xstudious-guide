package amazon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
)

func ConnectDB() *dynamodb.Client {

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	ddb_region := os.Getenv("AWS_REGION_DDB")

	ddbCfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(ddb_region),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		),
		//config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody), <- for debugging
	)
	ddbClient := dynamodb.NewFromConfig(ddbCfg)

	tables := map[string]func(*dynamodb.Client, string) error{
		"users": CreateUsersTable,
		"files": CreateFilesTable,
	}

	// Loop through tables
	for name, createFunc := range tables {
		err := CreateTableIfNotExists(createFunc, ddbClient, name)
		if err != nil {
			log.Fatalf("failed to create/check table %s: %v", name, err)
		}
		log.Printf("%s table ready for data\n", name)
	}

	log.Printf("Connected to DynamoDB\n")
	return ddbClient
}

func CreateTableIfNotExists(createFunc func(*dynamodb.Client, string) error, client *dynamodb.Client, tableName string) error {
	_, err := client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	if err == nil {
		log.Printf("%s table exists\n", tableName)
		return nil
	}

	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		return fmt.Errorf("error checking %s table: %w", tableName, err)
	}

	return createFunc(client, tableName)
}

func GetTables(client *dynamodb.Client) ([]string, error) {

	result, err := client.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if err != nil {
		return nil, err
	}
	return result.TableNames, nil
}

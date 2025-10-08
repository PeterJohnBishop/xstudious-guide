package amazon

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UserFile struct {
	UserID   string `dynamodbav:"userId"` // partition key
	FileID   string `dynamodbav:"fileId"` // sort key
	FileKey  string `dynamodbav:"fileKey"`
	Uploaded int64  `dynamodbav:"uploaded"`
}

func ConnectS3() (*s3.Client, string) {

	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	s3_region := os.Getenv("AWS_REGION_S3")

	s3Cfg, _ := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(s3_region),
		config.WithCredentialsProvider(
			aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		),
		//config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody), <- for debugging
	)
	s3Client := s3.NewFromConfig(s3Cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	_, err := s3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		msg := fmt.Sprintf("unable to load S3 buckets, %v", err)
		return nil, msg
	}
	log.Printf("Connected to S3\n")
	return s3Client, "Connected to S3"
}

func CreateFilesTable(client *dynamodb.Client, tableName string) error {
	_, err := client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("userId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("fileId"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("userId"),
				KeyType:       types.KeyTypeHash, // Partition key
			},
			{
				AttributeName: aws.String("fileId"),
				KeyType:       types.KeyTypeRange, // Sort key
			},
		},
		BillingMode: types.BillingModePayPerRequest, // On-demand billing
	})
	if err != nil {
		return fmt.Errorf("failed to create files table: %w", err)
	}

	fmt.Println("✅ Files table created:", tableName)
	return nil
}

func UploadFile(client *s3.Client, presigner *s3.PresignClient, filename string, fileContent multipart.File) (string, string, error) {
	bucketName := os.Getenv("AWS_BUCKET")

	fileKey := "uploads/" + filename

	_, err := client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
		Body:   fileContent,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to upload file: %w", err)
	}

	presignedReq, err := presigner.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(15*time.Minute)) // presigned URL valid for 15 minutes
	if err != nil {
		return "", "", fmt.Errorf("failed to generate presigned url: %w", err)
	}

	return fileKey, presignedReq.URL, nil
}

func SaveUserFile(dynamo *dynamodb.Client, tableName string, userFile UserFile) error {
	av, err := attributevalue.MarshalMap(userFile)
	if err != nil {
		return err
	}

	_, err = dynamo.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	return err
}

func GetUserFiles(dynamo *dynamodb.Client, userID string) ([]UserFile, error) {
	out, err := dynamo.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("files"),
		KeyConditionExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, err
	}

	var files []UserFile
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func DownloadFile(client *s3.Client, filename string) (string, error) {

	bucketName := os.Getenv("AWS_BUCKET")

	fileKey := filename

	expiration := time.Duration(5) * time.Minute

	presignClient := s3.NewPresignClient(client)
	presignedURL, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileKey),
	}, s3.WithPresignExpires(expiration))
	if err != nil {
		return "", fmt.Errorf("failed to presign URL: %w", err)
	}

	return presignedURL.URL, nil
}

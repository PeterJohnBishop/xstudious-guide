package amazon

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type User struct {
	ID       string `json:"id" dynamodbav:"id"`
	Name     string `json:"name" dynamodbav:"name"`
	Email    string `json:"email" dynamodbav:"email"`
	Password string `json:"password" dynamodbav:"password"`
}

func CreateUsersTable(client *dynamodb.Client, tableName string) error {
	_, err := client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash, // Primary Key
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("email-index"), // Name of the GSI
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("email"),
						KeyType:       types.KeyTypeHash, // Partition Key for GSI
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll, // include all attributes
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})
	return err
}

func CreateUser(client *dynamodb.Client, tableName string, item map[string]types.AttributeValue) error {
	email := item["email"].(*types.AttributeValueMemberS).Value

	queryOut, err := client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :e"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":e": &types.AttributeValueMemberS{Value: email},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return fmt.Errorf("failed to query GSI: %w", err)
	}

	if len(queryOut.Items) > 0 {
		return fmt.Errorf("user with email %s already exists", email)
	}

	_, err = client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func GetUserById(client *dynamodb.Client, tableName, id string) (map[string]types.AttributeValue, error) {
	result, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}
	return result.Item, nil
}

func GetAllUsers(client *dynamodb.Client, tableName string) ([]map[string]types.AttributeValue, error) {
	var items []map[string]types.AttributeValue
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		out, err := client.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:         aws.String(tableName),
			ExclusiveStartKey: lastEvaluatedKey,
		})
		if err != nil {
			return nil, err
		}

		items = append(items, out.Items...)

		if out.LastEvaluatedKey == nil {
			break
		}
		lastEvaluatedKey = out.LastEvaluatedKey
	}

	return items, nil
}

func GetUserByEmail(client *dynamodb.Client, tableName, email string) (*User, error) {
	email = strings.ToLower(email)

	result, err := client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, err
	}

	var user User
	err = attributevalue.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(client *dynamodb.Client, tableName string, user User) error {
	updateBuilder := expression.UpdateBuilder{}
	updatedFields := 0

	if user.Email != "" {
		expr, err := expression.NewBuilder().
			WithFilter(expression.Name("email").Equal(expression.Value(user.Email))).
			Build()
		if err != nil {
			return fmt.Errorf("error building email check expression: %w", err)
		}

		scanOut, err := client.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName:                 aws.String(tableName),
			FilterExpression:          expr.Filter(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})
		if err != nil {
			return fmt.Errorf("error checking for duplicate email: %w", err)
		}

		for _, item := range scanOut.Items {
			if idAttr, ok := item["id"].(*types.AttributeValueMemberS); ok {
				if idAttr.Value != user.ID {
					return fmt.Errorf("email %s is already in use", user.Email)
				}
			}
		}

		updateBuilder = updateBuilder.Set(expression.Name("email"), expression.Value(user.Email))
		updatedFields++
	}

	if user.Name != "" {
		updateBuilder = updateBuilder.Set(expression.Name("name"), expression.Value(user.Name))
		updatedFields++
	}

	if updatedFields == 0 {
		return fmt.Errorf("must update at least one field")
	}

	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		return fmt.Errorf("error in expression builder: %w", err)
	}

	_, err = client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: user.ID},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})
	if err != nil {
		return fmt.Errorf("error in client updater: %w", err)
	}

	return nil
}

func UpdatePassword(client *dynamodb.Client, tableName string, user User) error {
	updateBuilder := expression.UpdateBuilder{}
	updatedFields := 0

	if user.Name != "" && user.Email != "" && user.Password != "" {
		updateBuilder = updateBuilder.Set(expression.Name("password"), expression.Value(user.Password))
		updatedFields++
	}

	if updatedFields == 0 {
		fmt.Println("No fields to update")
		return fmt.Errorf("must update at least one field")
	}

	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		fmt.Println("Error in expression builder:", err)
		return err
	}

	_, err = client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: user.ID},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	})

	if err != nil {
		fmt.Println("Error in client updater:", err)
	}
	return err
}

func DeleteUser(client *dynamodb.Client, tableName, id string) error {
	_, err := client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	return err
}

package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

)

//CreateTable - test if it already exists first
func CreateTable()(*dynamodb.CreateTableOutput, error){

	config := &aws.Config{
		Region:   aws.String("us-west-2"),
		Endpoint: aws.String("http://localhost:8000"),
	}

	sess := session.Must(session.NewSession(config))

	svc := dynamodb.New(sess)

	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("uID"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("txID"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("uID"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("txID"),
				KeyType:       aws.String("RANGE"),
			},

		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String("Transactions"),
	}

	return svc.CreateTable(input)

}


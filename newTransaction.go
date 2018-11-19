package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

)

//Transaction - Sruct to compose transactions
type Transaction struct {
	UID string `json:"uID"`
	TXID string`json:"txID"`
}

//NewTransaction - Insert Transaction into Dynamo DB
func NewTransaction(tx Transaction){
	fmt.Printf("NewTransaction\n")

	config := &aws.Config{
		Region:   aws.String("us-west-2"),
		Endpoint: aws.String("http://localhost:8000"),
	}

	sess := session.Must(session.NewSession(config))

	svc := dynamodb.New(sess)

	info, err := dynamodbattribute.MarshalMap(tx)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal the movie, %v", err))
		}

	input := &dynamodb.PutItemInput{
			Item:      info,
			TableName: aws.String("Transactions"),
		}

		_, err = svc.PutItem(input)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

}
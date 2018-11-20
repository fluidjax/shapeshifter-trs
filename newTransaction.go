package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	log "github.com/Sirupsen/logrus"

)

//Transaction - Sruct to compose transactions
type Transaction struct {
	UID string `json:"uID"`
	TXID string`json:"txID"`
}

//NewTransaction - Insert Transaction into Dynamo DB
func NewTransaction(tx Transaction){

	sess, err := session.NewSession(&aws.Config{
        Region: aws.String("eu-west-2")},
    )

    // Create DynamoDB client
    svc := dynamodb.New(sess)

	info, err := dynamodbattribute.MarshalMap(tx)
		if err != nil {
			log.Fatal("failed to marshal the tx", err)
		}

	input := &dynamodb.PutItemInput{
			Item:      info,
			TableName: aws.String("Transactions"),
		}

	_, err = svc.PutItem(input)
		if err != nil {
			log.Warn(err.Error())
			return
		}
		log.Info("Created Item for ", tx.UID)
		return
}
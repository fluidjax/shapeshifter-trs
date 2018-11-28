package main

import (
	"encoding/hex"
	"errors"
	"strconv"

	// "strconv"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"

	log "github.com/Sirupsen/logrus"
)

//Transaction - Sruct to compose transactions
type Transaction struct {
	TxID          string    `json:"txID"`
	UserID        string    `json:"userID"`
	LeaderURL     string    `json:"leaderURL"`
	RingSignature string    `json:"ringSignature"`
	SignersCount  int       `json:"signersCount"`
	Signers       []int     `json:"signers"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Message       Message   `json:"message"`
	Policy        Policy    `json:"policy"`
}

//Message - details of the transaction
type Message struct {
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency"`
	DestinationAddress string `json:"destinationAddress"`
}

//Policy - who can sign
type Policy struct {
	Participants         []Participant `json:"participants"`
	NumberOfParticipants uint          `json:"numberOfParticipants"`
	Threshold            uint          `json:"threshold"`
}

//Participant - Details of signers
type Participant struct {
	URL      string `json:"url"`
	SK       string `json:"sK"`
	PK       string `json:"pk"`
	Approved bool   `json:"approved"`
}

//TransactionTableKey - find data in dynamodb
type TransactionTableKey struct {
	TxID string `json:"txID"`
}

//TransactionUpdate - date to update transaction
type TransactionUpdate struct {
	Approved bool  `json:":a"`
	Signer   []int `json:":s"`
}

//Generate pk and sk for all participants
func createRing(newTx Transaction) (tx Transaction, err error) {

	tx = newTx

	var params Parameters

	params.numberOfParticipants = tx.Policy.NumberOfParticipants
	params.threshold = tx.Policy.Threshold

	InitContext(params)

	for i := range tx.Policy.Participants {
		pK, sK := Keygen()
		tx.Policy.Participants[i].PK = hex.EncodeToString(pK)
		tx.Policy.Participants[i].SK = hex.EncodeToString(sK)
	}

	return tx, err

}

//CreateTransaction - Insert Transaction into Dynamo DB
func CreateTransaction(newTX Transaction) (tx Transaction, err error) {

	tx = newTX

	if newTX.TxID != "" {
		errStr := "Found TXID - Not a new transaction"
		return tx, errors.New(errStr)
	}

	txID, err := uuid.NewRandom()
	if err != nil {
		return tx, err
	}

	tx.TxID = txID.String()
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()
	tx.Policy.NumberOfParticipants = uint(len(tx.Policy.Participants))
	//I have to put a dummy value in the list... see below
	tx.Signers = append(tx.Signers, 0)
	tx, err = createRing(tx)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2")},
	)
	svc := dynamodb.New(sess)

	info, err := dynamodbattribute.MarshalMap(tx)
	if err != nil {
		return tx, err
	}

	input := &dynamodb.PutItemInput{
		Item:      info,
		TableName: aws.String("Transactions"),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return tx, err
	}

	log.Info("New TX added to database ", tx.TxID)

	return tx, err
}

//UpdateTransaction - Update transaction when signing approval is received
func UpdateTransaction(txID string, sr SignerRequest) (tx Transaction, err error) {

	config := &aws.Config{
		Region: aws.String("eu-west-2"),
	}

	sess := session.Must(session.NewSession(config))

	svc := dynamodb.New(sess)

	//Many hours wasted here - see https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/expression
	//this is wrong - it should be Add signers. to prevent me hainvg to put a dummy value in the slice
	update := expression.Set(
		expression.Name("policy.participants["+strconv.Itoa(sr.RingIndex)+"].approved"),
		expression.Value(true),
	).Set(
		expression.Name("signers"),
		expression.Name("signers").ListAppend(expression.Value([]int{sr.RingIndex})),
	).Set(
		expression.Name("signersCount"),
		expression.Name("signersCount").Plus(expression.Value(1)),
	)

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Key: map[string]*dynamodb.AttributeValue{
			"txID": {
				S: aws.String(sr.TxID),
			},
		},
		TableName:           aws.String("Transactions"),
		ReturnValues:        aws.String("ALL_NEW"),
		ConditionExpression: aws.String("policy.threshold > signersCount"),
		UpdateExpression:    expr.Update(),
	}

	result, err := svc.UpdateItem(input)

	dynamodbattribute.UnmarshalMap(result.Attributes, &tx)

	return tx, err

	// var txupdatedata TransactionUpdate
	// txupdatedata.Approved = true
	// txupdatedata.Signer = make([]int, sr.RingIndex)
	// updateData, err := dynamodbattribute.MarshalMap(txupdatedata)
	// if err != nil {
	// 	log.Info("Marshall map", err.Error())
	// 	return
	// }

	// updateString := "set policy.participants[" + strconv.Itoa(sr.RingIndex) + "].approved=:a signers=:s"
	// updateString := "set signers = :s"
	// input := &dynamodb.UpdateItemInput{
	// 	Key: map[string]*dynamodb.AttributeValue{
	// 		"txID": {
	// 			S: aws.String(sr.TxID),
	// 		},
	// 	},
	// 	TableName:                 aws.String("Transactions"),
	// 	UpdateExpression:          aws.String(updateString),
	// 	ExpressionAttributeValues: updateData,

	// 	ReturnValues: aws.String("ALL_NEW"),
	// }

}

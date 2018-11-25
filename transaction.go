package main

import (
	"encoding/hex"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"

	log "github.com/Sirupsen/logrus"
)

//Transaction - Sruct to compose transactions
type Transaction struct {
	TxID          string    `json:"txID"`
	UserID        string    `json:"userID"`
	LeaderURL     string    `json:"leaderURL"`
	RingSignature string    `json:"ringSignature"`
	Signers       []int64   `json:"signers"`
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

//ApproveTransaction - Update transaction when signing approval is received
func ApproveTransaction(txID string, sr SignerRequest) (tx Transaction, err error) {

	log.Info("Going to update: ", txID)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-2")},
	)
	svc := dynamodb.New(sess)

	//hard bit goes here

	input := &dynamodb.UpdateItemInput{}

	_, err = svc.UpdateItem(input)

	if err != nil {
		log.Warn(err.Error())
	}

	return tx, err

}

package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

//ApprovalRequest - Similar to Transaction
type ApprovalRequest struct {
	TxID      string    `json:"txID"`
	UserID    string    `json:"userID"`
	CreatedAt time.Time `json:"createdAt"`
	Message   Message   `json:"message"`
	URL       string    `json:"url"`
	RingIndex int       `json:"ringIndex"`
	Approval  bool      `json:"approval"`
	PK        string    `json:"pk"`
	SK        string    `json:"sk"`
}

// SignatureRequest - forming the ring
type SignatureRequest struct {
	TxID       string    `json:"txID"`
	UserID     string    `json:"userID"`
	RingIndex  int       `json:"PartIndex"`
	CreatedAt  time.Time `json:"createdAt"`
	Message    Message   `json:"message"`
	Signers    []uint    `json:"signers"`
	PublicKeys []string  `json:"publicKeys"`
	SK         string    `json:"sk"`
	PSig       string    `json:"pSig"`
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handleTransaction(w http.ResponseWriter, r *http.Request) {

	var tx Transaction
	var err error

	if r.Method == http.MethodPost {
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&tx)

		//TODO: check there is only one leader

		tx, err = CreateTransaction(tx)
		if err != nil {
			log.Warn(err)
		}

		go setUpApprovalRequest(tx)

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func setUpApprovalRequest(tx Transaction) {

	var approvalRequest ApprovalRequest

	approvalRequest.TxID = tx.TxID
	approvalRequest.UserID = tx.UserID
	approvalRequest.CreatedAt = time.Now()
	approvalRequest.Message = tx.Message

	for i, p := range tx.Policy.Participants {
		approvalRequest.URL = tx.Policy.Participants[i].URL
		approvalRequest.RingIndex = i

		if p.Leader {
			StoreApproval(approvalRequest)
		} else {
			go postApprovalRequest(p.URL, approvalRequest)
		}
	}
}

func postApprovalRequest(url string, approvalRequest ApprovalRequest) {

	qpprovalRequestJSON, err := json.Marshal(approvalRequest)
	if err != nil {
		log.Warn(err)
	}
	resp, err := http.Post(url+"/approvalrequest", "application/json", bytes.NewBuffer(qpprovalRequestJSON))
	if err != nil {
		log.Warn(err)
	}

	var approvalRequestResponse ApprovalRequest

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&approvalRequestResponse)
	if err != nil {
		log.Warn(err)
	}

	//this needs to be refactored maybe
	updatedTX, err := StoreApproval(approvalRequestResponse)
	if err != nil {

		if strings.Contains(err.Error(), "ConditionalCheckFailedException:") {
			log.Info("Got enough signers already thanks!")
		} else {
			log.Warn(err)
		}
	} else {

		var approvals uint

		for _, p := range updatedTX.Policy.Participants {
			if p.Approved {
				approvals++
			}
		}

		if approvals < updatedTX.Policy.Threshold {
			log.Info("Wating for more approvals")
		}
		if approvals == updatedTX.Policy.Threshold {
			go setUpSignatures(updatedTX)
		}
		if approvals > updatedTX.Policy.Threshold {
			log.Info("Don't need this one")
		}
	}

	// mJSON, _ := json.MarshalIndent(sigRequestResponse, "", "\t")
	// fmt.Printf("%s\n", mJSON)

}

func handleApprovalRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		var approvalRequest ApprovalRequest
		var err error

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&approvalRequest)

		log.Info("Got Signing Request for ", approvalRequest.TxID)
		if err != nil {
			log.Warn(err)
		}

		approvalInterval := rand.Intn(5) * 1000
		time.Sleep(time.Duration(approvalInterval) * time.Millisecond)

		approvalRequest.Approval = true

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(approvalRequest)

	}

}

func setUpSignatures(tx Transaction) {

	log.Info("Got enough approvals, let's go!")

	var signersSlice []uint
	var publicKeySlice []string

	for i, p := range tx.Policy.Participants {

		publicKeySlice = append(publicKeySlice, tx.Policy.Participants[i].PK)

		if p.Approved {
			signersSlice = append(signersSlice, uint(i))
		}
	}

	for i, p := range tx.Policy.Participants {

		if p.Approved {
			var partSign SignatureRequest
			partSign.TxID = tx.TxID
			partSign.UserID = tx.UserID
			partSign.RingIndex = i
			partSign.Message = tx.Message
			partSign.CreatedAt = time.Now()
			partSign.SK = p.SK
			partSign.Signers = signersSlice
			partSign.PublicKeys = publicKeySlice

			go postSignatureRequest(partSign, tx.Policy.Participants[i].URL)
		}
	}
}

func postSignatureRequest(sigRequest SignatureRequest, url string) {

	sigRequestJSON, err := json.Marshal(sigRequest)
	if err != nil {
		log.Warn(err)
	}
	resp, err := http.Post(url+"/signaturerequest", "application/json", bytes.NewBuffer(sigRequestJSON))
	if err != nil {
		log.Warn(err)
	}

	var sigRequestResponse SignatureRequest

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&sigRequestResponse)
	if err != nil {
		log.Warn(err)
	}

	tx, err := StorePSig(sigRequestResponse)
	if err != nil {
		log.Warn(err)
	}

	sigCount := 0
	for i, p := range tx.Policy.Participants {
		if p.PSig != "" {
			fmt.Printf("Got sig for \t%v\n", i)
			sigCount++
		}
		if uint(sigCount) == tx.Policy.Threshold {
			go leaderSign(tx)
		}
	}
}

func handleSignatureRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		var sigDetails SignatureRequest
		var err error

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&sigDetails)
		if err != nil {
			log.Warn("here ", err)
		}

		log.Info("Got Signing Request for ", sigDetails.TxID)

		//turn the message into a byte array
		encBuf := new(bytes.Buffer)
		err = gob.NewEncoder(encBuf).Encode(sigDetails.Message)
		if err != nil {
			log.Warn(err)
		}
		message := encBuf.Bytes()

		privateKey, err := hex.DecodeString(sigDetails.SK)
		var pubKeyBytes []byte

		for _, p := range sigDetails.PublicKeys {
			pByte, _ := hex.DecodeString(p)
			pubKeyBytes = append(pubKeyBytes, pByte...)
		}

		params := Parameters{uint(len(sigDetails.PublicKeys)), uint(len(sigDetails.Signers))}

		InitContext(params)

		pSig := ParticipantSign(message, privateKey, sigDetails.Signers, pubKeyBytes)

		sigDetails.PSig = hex.EncodeToString(pSig)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sigDetails)

	}
}

func leaderSign(tx Transaction) {

	//turn the message into a byte array
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(tx.Message)
	if err != nil {
		log.Warn(err)
	}
	message := encBuf.Bytes()

	//find the leader and signers

	var leaderIndex uint
	var signers []uint
	var privateKey []byte
	var publicKeys []byte
	var pSigs []byte

	for i, p := range tx.Policy.Participants{
		if p.Leader{
			leaderIndex=uint(i)
			privateKey, _ = hex.DecodeString(p.SK)
		}
		if p.PSig != ""{
			signers = append(signers, uint(i))

			pSig, _ := hex.DecodeString(p.PSig)
			pSigs = append(pSigs, pSig...)
		}
		pk, _ := hex.DecodeString(p.PK)
		publicKeys = append(publicKeys, pk...)

	}

	params := Parameters{uint(len(tx.Policy.Participants)), uint(tx.Policy.Threshold)}

	InitContext(params)

	ringSig:= leader_sign(message, leaderIndex, privateKey, signers, publicKeys, pSigs)

	fmt.Printf("ringSig %v", ringSig)

}

func main() {

	serverPort := flag.Int("port", 5000, "Server Port")

	flag.Parse()

	port := ":" + strconv.Itoa(*serverPort)
	log.Info("Server Port = ", port)

	http.HandleFunc("/transaction/", handleTransaction)
	http.HandleFunc("/approvalrequest", handleApprovalRequest)
	http.HandleFunc("/signaturerequest", handleSignatureRequest)

	if err := http.ListenAndServe(port, logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

//SignerRequest - Similar to Transaction
type SignerRequest struct {
	TxID      string    `json:"txID"`
	UserID    string    `json:"userID"`
	LeaderURL string    `json:"leaderURL"`
	CreatedAt time.Time `json:"createdAt"`
	Message   Message   `json:"message"`
	URL       string    `json:"url"`
	PublicKey string    `json:"publicKey"`
	Approval  bool      `json:"approval"`
}

func inviteSigners(tx Transaction) {

	var signRequest SignerRequest

	signRequest.TxID = tx.TxID
	signRequest.UserID = tx.UserID
	signRequest.LeaderURL = tx.LeaderURL
	signRequest.CreatedAt = time.Now()
	signRequest.Message = tx.Message

	for i, p := range tx.Policy.Participants {
		// log.Info("Inviting ", tx.Policy.Participants[i].URL)
		signRequest.URL = tx.Policy.Participants[i].URL
		signRequest.PublicKey = tx.Policy.Participants[i].PK

		signRequestJSON, err := json.Marshal(signRequest)
		if err != nil {
			log.Fatalln(err)
		}

		_, err = http.Post(p.URL+"/signingrequest", "application/json", bytes.NewBuffer(signRequestJSON))
		if err != nil {
			log.Warn(err)
		}
		//TODO: Need to handle closing the request - throws an error on 404
		// defer resp.Body.Close()

	}
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

	switch r.Method {

	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&tx)
		tx, err = CreateTransaction(tx)
		if err != nil {
			log.Warn(err)
		}

		go inviteSigners(tx)

	case http.MethodPut:
		tx := r.URL.Path[len("/transaction/"):]
		var approvedRequest SignerRequest

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&approvedRequest)

		mJSON, _ := json.MarshalIndent(approvedRequest, "", "\t")
		fmt.Printf("%s\n", mJSON)

		updatedTX, err := ApproveTransaction(tx, approvedRequest)
		if err != nil {
			log.Warn(err)
		}

		mJSON, _ = json.MarshalIndent(updatedTX, "", "\t")
		fmt.Printf("%s\n", mJSON)

	case http.MethodDelete:

	default:

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func handleSigningRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		var signingRequest SignerRequest
		var err error

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&signingRequest)

		log.Info("Got Signing Request for ", signingRequest.TxID)
		if err != nil {
			log.Warn(err)
		}

		w.WriteHeader(200)

		//Insert logic for approving message
		approveMessage(true, signingRequest)

	case http.MethodPut:

	case http.MethodDelete:

	default:

	}

}

func approveMessage(approval bool, signingRequest SignerRequest) {

	if approval {
		//Add a delay to make a better demo
		approvalInterval := rand.Intn(5) * 1000
		time.Sleep(time.Duration(approvalInterval) * time.Millisecond)

		signingRequest.Approval = true
		log.Info("Message approved")

		url := signingRequest.LeaderURL + "/transaction/" + signingRequest.TxID

		signingRequestJSON, err := json.Marshal(signingRequest)
		if err != nil {
			log.Warn(err)
		}

		//No convenience method for PUT so having to do this:
		client := &http.Client{}
		request, err := http.NewRequest("PUT", url, bytes.NewBuffer(signingRequestJSON))
		request.Header.Add("Content-Type", "application/json")
		resp, err := client.Do(request)
		if err != nil {
			log.Warn(err)
		}

		//TODO: this is going to cause a crash on 404
		defer resp.Body.Close()
	}
}
func main() {

	http.HandleFunc("/transaction/", handleTransaction)
	http.HandleFunc("/signingrequest", handleSigningRequest)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

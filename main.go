package main

import (
	"bytes"
	"encoding/json"
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
	RingIndex int       `json:"ringIndex"`
	Approval  bool      `json:"approval"`
	PK        string    `json:"pk"`
	SK        string    `json:"sk"`
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

		go inviteApprovers(tx)

	case http.MethodPut:
		tx := r.URL.Path[len("/transaction/"):]
		var approvedRequest SignerRequest

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&approvedRequest)

		updatedTX, err := UpdateTransaction(tx, approvedRequest)
		if err != nil {
			log.Warn(err)
		}

		var approvals uint

		for _, p := range updatedTX.Policy.Participants {
			if p.Approved {
				approvals++
			}
		}

	case http.MethodDelete:

	default:

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func inviteApprovers(tx Transaction) {

	var approvalRequest SignerRequest

	approvalRequest.TxID = tx.TxID
	approvalRequest.UserID = tx.UserID
	approvalRequest.LeaderURL = tx.LeaderURL
	approvalRequest.CreatedAt = time.Now()
	approvalRequest.Message = tx.Message

	for i, p := range tx.Policy.Participants {
		// log.Info("Inviting ", tx.Policy.Participants[i].URL)
		approvalRequest.URL = tx.Policy.Participants[i].URL
		approvalRequest.RingIndex = i

		go postApprovalRequest(p.URL, approvalRequest)

	}
}

func postApprovalRequest(url string, approvalRequest SignerRequest) {

	qpprovalRequestJSON, err := json.Marshal(approvalRequest)
	if err != nil {
		log.Warn(err)
	}
	_, err = http.Post(url+"/approvalrequest", "application/json", bytes.NewBuffer(qpprovalRequestJSON))
	if err != nil {
		log.Warn(err)
	}

	// if approvals == updatedTX.Policy.Threshold {
	// 	requestSignatures(updatedTX)
	// } else {
	// 	log.Info("Not enough signers yet")
	// }
	//TODO: Need to handle closing the request - throws an error on 404
	// defer resp.Body.Close()
}

func handleApprovalRequest(w http.ResponseWriter, r *http.Request) {
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
		log.Info("Message approved, gonna respond to: ", signingRequest.LeaderURL)

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

func requestSignatures(tx Transaction) {

	log.Info("We have enough signers")
	// for i, p := range tx.Policy.Participants {

	// 	var sigRequest Transaction
	// 	sigRequest.TxID = tx.TxID
	// 	sigRequest.LeaderURL = tx.LeaderURL
	// 	sigRequest.Message = tx.Message
	// 	// go postSigRequest()
	// }
}

func postSigRequest(sigRequest SignerRequest) {
	// http.Post()
}

func handleSign(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		log.Info("Handle Sign")
	}
}

func main() {

	http.HandleFunc("/transaction/", handleTransaction)
	http.HandleFunc("/approvalrequest", handleApprovalRequest)
	http.HandleFunc("/participantsign", handleSign)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}
}

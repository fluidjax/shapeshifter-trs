package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
	"math/rand"

	log "github.com/Sirupsen/logrus"
)

//SignerRequest - Similar to Transaction
type SignerRequest struct {
	TxID      string    `json:"txID"`
	UserID    string    `json:"userID"`
	CreatedAt time.Time `json:"createdAt"`
	Message   Message   `json:"message"`
}

func inviteSigners(tx Transaction) {

	var signRequest SignerRequest

	signRequest.TxID = tx.TxID
	signRequest.UserID = tx.UserID
	signRequest.CreatedAt = time.Now()
	signRequest.Message = tx.Message

	signRequestJSON, err := json.Marshal(signRequest)
	if err != nil {
		log.Fatalln(err)
	}

	for i, p := range tx.Policy.Participants {
		log.Info("Inviting ", i)
		_, err = http.Post(p.URL+"/signingrequest", "application/json", bytes.NewBuffer(signRequestJSON))
		if err != nil {
			log.Warn(err)
		}
		
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

	case http.MethodDelete:

	default:

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func handleSigningRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		var signingRequest Transaction
		var err error

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(&signingRequest)

		log.Info("Got Signing Request for ", signingRequest.TxID)
		if err != nil {
			log.Warn(err)
		}

		w.WriteHeader(200)

		//Insert logic for approving message
		approveMessage(true)


	case http.MethodPut:

	case http.MethodDelete:

	default:

	}

}

func approveMessage(approval bool){

	if approval {
		approvalInterval := rand.Intn(5) * 1000
		time.Sleep(time.Duration(approvalInterval) * time.Millisecond)
		log.Info("Message approved")
	}
}
func main() {

	http.HandleFunc("/transaction", handleTransaction)
	http.HandleFunc("/signingrequest", handleSigningRequest)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}

}

// //ServerConfig - config details for this sever
// var ServerConfig Config

// func handleTransaction(w http.ResponseWriter, r *http.Request) {

// 	switch r.Method {
// 	case http.MethodGet:
// 		// Serve the resource.
// 	case http.MethodPost:
// 		transaction := CreateTransaction()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(transaction)
// 	case http.MethodPut:
// 		// Update an existing record.
// 	case http.MethodDelete:
// 		// Remove the record.
// 	default:
// 		// Give an error message.
// 	}
// }

// func handleServerConfig(w http.ResponseWriter, r *http.Request) {

// 	switch r.Method {

// 	case http.MethodPost:
// 		decoder := json.NewDecoder(r.Body)
// 		err := decoder.Decode(&ServerConfig)
// 		if err != nil {
// 			log.Warn(err)
// 		}
// 		go pollForTransactions()
// 	case http.MethodPut:
// 		// Update an existing record.
// 	case http.MethodDelete:
// 		ServerConfig.DirectoryServer = ""
// 	default:
// 		// Give an error message.
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(ServerConfig)

// }

// func handleGetKeys(w http.ResponseWriter, r *http.Request) {

// 	// p := parameters{number_of_participants: 10, threshold: 5}
// 	// InitContext(p)

// 	// pK, sK := Keygen()

// 	// ServerConfig.TRSPublicKey = hex.EncodeToString(pK)
// 	// ServerConfig.TRSPrivateKey = hex.EncodeToString(sK)

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(ServerConfig)

// }

// func pollForTransactions() {

// 	directoryString := "http://" + ServerConfig.DirectoryServer + "/transaction"

// 	for ServerConfig.DirectoryServer != "" {
// 		txResp, err := http.Get(directoryString)
// 		if err != nil {
// 			log.Warn(err)
// 		} else {
// 			fmt.Printf("txResp %v", &txResp)
// 		}
// 		pollInterval := rand.Intn(5) * 1000
// 		time.Sleep(time.Duration(pollInterval) * time.Millisecond)
// 	}
// }

// func main() {

// 	http.HandleFunc("/serverconfig", handleServerConfig)
// 	http.HandleFunc("/keys", handleGetKeys)
// 	http.HandleFunc("/transaction", handleTransaction)

// 	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
// 		panic(err)
// 	}

// }

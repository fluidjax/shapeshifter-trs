package main

import (
	"encoding/json"
	// "fmt"

	// "encoding/hex"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

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
		if err != nil{
			log.Warn(err)
		}

	case http.MethodPut:
		
	case http.MethodDelete:
		
	default:
		
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func main() {

	http.HandleFunc("/transaction", handleTransaction)

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

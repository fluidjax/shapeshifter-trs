package main

import (
	"encoding/json"
	"fmt"

	// "encoding/hex"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

//Config - server setup details
type Config struct {
	DirectoryServer string `json:"directoryServer"`
}

//ServerConfig - config details for this sever
var ServerConfig Config

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func handleTransaction(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodGet:
		// Serve the resource.
	case http.MethodPost:
		transaction := CreateTransaction()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transaction)
	case http.MethodPut:
		// Update an existing record.
	case http.MethodDelete:
		// Remove the record.
	default:
		// Give an error message.
	}
}

func getStatus(w http.ResponseWriter, r *http.Request) {

	if ServerConfig.DirectoryServer != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ServerConfig)
	} else {
		w.Write([]byte("Server Not Yet Confgured POST Directory Server URL To /setdirectory"))
	}

}

func handleDirectory(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("handleDirectory\n")

	switch r.Method {
	case http.MethodGet:
		// Serve the resource.
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&ServerConfig)
		if err != nil {
			log.Warn(err)
		}
	case http.MethodPut:
		// Update an existing record.
	case http.MethodDelete:
		ServerConfig.DirectoryServer = ""
	default:
		// Give an error message.
	}

	confJSON, _ := json.MarshalIndent(ServerConfig, "", "\t")
	fmt.Printf("%s\n", confJSON)

	go pollForTransactions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ServerConfig)
}

func handleGetKeys(w http.ResponseWriter, r *http.Request) {

	// p := parameters{number_of_participants: 10, threshold: 5}
	// InitContext(p)

	// pK, sK := Keygen()

	// ServerConfig.TRSPublicKey = hex.EncodeToString(pK)
	// ServerConfig.TRSPrivateKey = hex.EncodeToString(sK)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ServerConfig)

}

func pollForTransactions() {

	directoryString := "http://" + ServerConfig.DirectoryServer + "/transaction"

	for ServerConfig.DirectoryServer != "" {
		txResp, err := http.Get(directoryString)
		if err != nil {
			log.Warn(err)
		} else {
			fmt.Printf("txResp %v", &txResp)
		}
		time.Sleep(2000 * time.Millisecond)
	}
}

func main() {

	http.HandleFunc("/", getStatus)
	http.HandleFunc("/directory", handleDirectory)
	http.HandleFunc("/keys", handleGetKeys)
	http.HandleFunc("/transaction", handleTransaction)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}

}

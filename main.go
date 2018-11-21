package main

import (
	"encoding/json"
	"fmt"

	// "encoding/hex"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
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

func newTxHandler(w http.ResponseWriter, r *http.Request) {
	uID, err := uuid.NewRandom()
	if err != nil {
		log.Warn(err)
	}

	txID, err := uuid.NewRandom()
	if err != nil {
		log.Warn(err)
	}

	tx := Transaction{uID.String(), txID.String()}

	NewTransaction(tx)

}

func getStatus(w http.ResponseWriter, r *http.Request) {

	if ServerConfig.DirectoryServer != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ServerConfig)
	} else {
		w.Write([]byte("Server Not Yet Confgured POST Directory Server URL To /setdirectory"))
	}

}

func handleSetDirectory(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&ServerConfig)
	if err != nil {
		log.Warn(err)
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

func pollForTransactions(){

	directoryString := "http://" + ServerConfig.DirectoryServer + "/getTransactions"

	for{
	txResp, err := http.Get(directoryString)
	if err != nil {
		log.Warn(err)
	}else {
		fmt.Printf("txResp %v", &txResp)
	}
	time.Sleep(2000 * time.Millisecond)
	}
}

func main() {

	http.HandleFunc("/", getStatus)
	http.HandleFunc("/setdirectory", handleSetDirectory)
	http.HandleFunc("/getkeys", handleGetKeys)
	http.HandleFunc("/newtransaction", newTxHandler)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}

}

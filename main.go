package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
)

func newTxHandler(w http.ResponseWriter, r *http.Request){
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
	
	fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {

	http.HandleFunc("/", getStatus)
	http.HandleFunc("/newtransaction", newTxHandler)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}

}

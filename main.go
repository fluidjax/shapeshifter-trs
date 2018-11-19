package main

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
)

func init() {

	//Need to test if the table exists before running this
	result, err := CreateTable()

	if err != nil {
		log.Info("DB Table Is Set Up")
	} else {
		log.Info(result)
	}

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

	uID, err := uuid.NewRandom()
	if err != nil {
		log.Warn(err)
	} else {
		fmt.Println(uID)
	}

	txID, err := uuid.NewRandom()
	if err != nil {
		log.Warn(err)
	} else {
		fmt.Println(txID)
	}

	tx := Transaction{uID.String(), txID.String()}

	NewTransaction(tx)

	TrsTest()

	http.HandleFunc("/", getStatus)

	if err := http.ListenAndServe(":5000", logRequest(http.DefaultServeMux)); err != nil {
		panic(err)
	}

}

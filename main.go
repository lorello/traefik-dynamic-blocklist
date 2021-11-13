// https://medium.com/the-andela-way/build-a-restful-json-api-with-golang-85a83420c9da
// 
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type ip struct {
	ID          string `json:"ID"`
	Description string `json:"Description"`
	Ip	    string `json:"Ip"`
}

type allIps map[string]ip


var ips = map[string]ip{
	"192.168.1.1": {
		ID:		"1",
		Description:	"Revelaed from loki logs @ 2010291029120",
		Ip:		"192.168.178.1",
		},
	}


func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func createIp(w http.ResponseWriter, r *http.Request) {
	var newIp ip
	// Convert r.Body into a readable formart
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event id, title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newIp)

	// Add the newly created event to the array of ips
	ips = append(ips, newIp)

	// Return the 201 created status code
	w.WriteHeader(http.StatusCreated)
	// Return the newly created event
	json.NewEncoder(w).Encode(newIp)
}


func isAllowed(w http.ResponseWriter, r *http.Request) {

    ip := r.RemoteAddr
    xforward := r.Header.Get("X-Forwarded-For")
    fmt.Println("IP: ", ip)
    fmt.Println("X-Forwarded-For: ", xforward)

	// Get the details from an existing event
	// Use the blank identifier to avoid creating a value that will not be used
	for _, singleIP := range ips {
		if singleIP.Ip == ip {
			json.NewEncoder(w).Encode(singleIP)
		}
		if singleIP.Ip == xforward {
			json.NewEncoder(w).Encode(singleIP)
		}

	}
}


func getBlacklistItem(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the url
	ipID := mux.Vars(r)["id"]

	// Get the details from an existing event
	// Use the blank identifier to avoid creating a value that will not be used
	for _, singleIP := range ips {
		if singleIP.ID == ipID {
			json.NewEncoder(w).Encode(singleIP)
		}
	}
}

func getBlacklist(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ips)
}

func updateBlacklistItem(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the url
	ipID := mux.Vars(r)["id"]
	var updatedEvent ip
	// Convert r.Body into a readable formart
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the event title and description only in order to update")
	}

	json.Unmarshal(reqBody, &updatedEvent)

	for i, singleIP := range ips {
		if singleIP.ID == ipID {
			singleIP.Description = updatedEvent.Description
			ips[i] = singleIP
			json.NewEncoder(w).Encode(singleIP)
		}
	}
}

func deleteBlacklistItem(w http.ResponseWriter, r *http.Request) {
	// Get the ID from the url
	ipID := mux.Vars(r)["id"]

	// Get the details from an existing event
	// Use the blank identifier to avoid creating a value that will not be used
	for i, singleIP := range ips {
		if singleIP.ID == ipID {
			ips = append(ips[:i], ips[i+1:]...)
			fmt.Fprintf(w, "The event with ID %v has been deleted successfully", ipID)
		}
	}
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/check", isAllowed).Methods("GET")
	router.HandleFunc("/blacklist", createIp).Methods("POST")
	router.HandleFunc("/blacklist", getBlacklist).Methods("GET")
	router.HandleFunc("/blacklist/{id}", getBlacklistItem).Methods("GET")
	router.HandleFunc("/blacklist/{id}", updateBlacklistItem).Methods("PATCH")
	router.HandleFunc("/blacklist/{id}", deleteBlacklistItem).Methods("DELETE")
	fmt.Println("Server started, listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}



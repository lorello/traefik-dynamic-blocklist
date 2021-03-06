// https://medium.com/the-andela-way/build-a-restful-json-api-with-golang-85a83420c9da

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"time"
	"viper"
)

type ip struct {
	Description string `json:"Description"`
	Ip          string `json:"Ip"`
	// TODO: Expire time
}



// The structure that will host the blocked IPs
var ips = map[string]ip{

	// Todo: remove this hardcoded value, added for test
	"192.168.1.1": {
		Description: "Revelaed from loki logs @ 2010291029120",
		Ip:          "192.168.178.1",
	},
}

var suspectsLimit int = 5
var suspected = map[string]int{}

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This service should not be call directly!")
}

// Add blocked IPs with an API call
func createIp(w http.ResponseWriter, r *http.Request) {
	var newIp ip
	// Convert r.Body into a readable formart
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Kindly enter data with the IP id, title and description only in order to update")
	}

	json.Unmarshal(reqBody, &newIp)

	// Add the newly created IP to the array of ips
	ips[newIp.Ip] = newIp

	// Return the 201 created status code
	w.WriteHeader(http.StatusCreated)
	// Return the newly created IP
	json.NewEncoder(w).Encode(newIp)
}

// This api should receive calls from traefik
// ForwardAuth Middleware:
// https://doc.traefik.io/traefik/v2.5/middlewares/http/forwardauth/
// Check if an IP should be allowed or not
// and at the same time check if the path requested
// could be a symptom of an attack
func isAllowed(w http.ResponseWriter, r *http.Request) {

	// get request data from Traefik
	originIp := r.RemoteAddr
	xforward := r.Header.Get("X-Forwarded-For")
	if xforward != "" {
		originIp = xforward
	}
	originHost := r.Header.Get("X-Forwarded-Host")
	originUri := r.Header.Get("X-Forwarded-Uri")

        var result bool = false

	val, result := ips[originHost]

	if result {
		log.Printf("%d. Blocked request on %s%s from %s (%s)", counter, originHost, originUri, originIp, val.Description)
		http.Error(w, "Forbidden", http.StatusForbidden)
	} else {
		// log.Printf("%d. Allowed request on %s%s from %s", counter, originHost, originUri, originIp)

		// check if host is attacking
		// TODO: other checks to implement
		// cgi-bin, wp-admin
		// Good source of check 
		// https://github.com/PHPIDS/PHPIDS/blob/master/lib/IDS/default_filter.xml
		// https://github.com/jensvoid/lorg
		// All this thread is enough: https://security.stackexchange.com/questions/200427/what-are-the-common-features-to-identify-xss-attack-from-apache-log-file
		// all these check should become async, so put it in a goroutine
		listOfPatterns := []string{"wp-admin", "cgi-bin", "phpMyAdmin", "jndi"}
		for _, regex := range listOfPatterns {
			didMatched, err := regexp.MatchString(regex, originUri)
			if err != nil {
				log.Fatal(err)
			}
			if didMatched == true {
				_, result2 := suspected[originHost]
				if result2 {
					suspected[originHost]+=1
				} else {
					suspected[originHost]=1
				}
				if suspected[originHost] >= suspectsLimit {

					newIp := ip{
						Description: "Added after 5 suspects",
						Ip:          originHost,
					}
					ips[newIp.Ip] = newIp

					log.Printf("%d. Maximum suspects reached on %s%s from %s", counter, originHost, originUri, originIp)
				} else {
					log.Printf("%d. Attack suspect on %s%s from %s", counter, originHost, originUri, originIp)
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}

func getBlacklist(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(ips)
}

var counter = 0

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get metrics here
		//log.Println(r.RequestURI)
		counter += 1
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func main() {
	var port int
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.IntVar(&port, "port", 8000, "HTTP server port")
	flag.Parse()


	viper.SetDefault("myListOfPatterns", [1]string{"cgi-bin"})
	viper.BindEnv("myListOfPatterns")

	// Mux lib: https://github.com/gorilla/mux
	router := mux.NewRouter().StrictSlash(true)

	router.Use(loggingMiddleware)

	router.HandleFunc("/", homeLink)
	router.HandleFunc("/check", isAllowed).Methods("GET")
	router.HandleFunc("/blacklist", createIp).Methods("POST")
	//router.HandleFunc("/blacklist", getBlacklist).Methods("GET")

	bindAddr := fmt.Sprint("0.0.0.0:", port)
	srv := &http.Server{
		Addr: bindAddr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 10,
		Handler:      router, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		log.Println("Server started, listening on port", port)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	//go func() {
	//	log.Println(len(ips))
	//}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

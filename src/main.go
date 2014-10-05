package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

var siteInfo SiteInfo

func main() {
	// Command line parameters
	bind := flag.String("port", ":8080", "Address to bind to")
	mongo := flag.String("dburl", "localhost", "MongoDB servers, separated by comma")
	dbname := flag.String("dbname", "maud", "MongoDB database to use")
	flag.Parse()

	// Load Site info file
	rawconf, _ := ioutil.ReadFile("info.json")
	err := json.Unmarshal(rawconf, &siteInfo)
	if err != nil {
		panic(err)
	}

	// Initialize database
	DBInit(*mongo, *dbname)
	defer DBClose()

	// Setup request handlers
	router := mux.NewRouter()
	GET := router.Methods("GET").Subrouter()
	POST := router.Methods("POST").Subrouter()

	GET.HandleFunc("/", httpHome)
	GET.HandleFunc("/tag/{tag}", httpTagSearch)
	GET.HandleFunc("/thread/{thread}", httpThread)
	GET.HandleFunc("/thread/{thread}/{page}", httpThread)
	GET.HandleFunc("/new", httpNewThread)

	POST.HandleFunc("/new", apiNewThread)
	POST.HandleFunc("/t{thread}/reply", apiReply)

	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start serving pages
	fmt.Printf("Listening on %s\r\n", *bind)
	http.ListenAndServe(*bind, nil)
}

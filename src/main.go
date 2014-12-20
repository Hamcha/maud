package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var siteInfo SiteInfo

// absolute path to Maud root directory
var maudRoot string

func main() {
	// get executable path
	maudExec, err := filepath.Abs(os.Args[0])
	if err != nil {
		panic(err)
	}
	maudRoot = filepath.Dir(maudExec)

	// Command line parameters
	bind := flag.String("port", ":8080", "Address to bind to")
	mongo := flag.String("dburl", "localhost", "MongoDB servers, separated by comma")
	dbname := flag.String("dbname", "maud", "MongoDB database to use")
	flag.StringVar(&maudRoot, "root", maudRoot, "The HTTP server root directory")
	flag.Parse()

	// Load Site info file
	rawconf, _ := ioutil.ReadFile(maudRoot + "/info.json")
	err = json.Unmarshal(rawconf, &siteInfo)
	if err != nil {
		panic(err)
	}

	// Initialize parsers
	initMarkdown()
	initbbcode()

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
	GET.HandleFunc("/thread/{thread}/page/{page}", httpThread)
	GET.HandleFunc("/new", httpNewThread)

	POST.HandleFunc("/new", apiNewThread)
	POST.HandleFunc("/thread/{thread}/reply", apiReply)
	POST.HandleFunc("/thread/{thread}/post/{post}/edit", apiEditPost)

	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(maudRoot+"/static"))))

	// Start serving pages
	fmt.Printf("Listening on %s\r\nServer root: %s\r\n", *bind, maudRoot)
	http.ListenAndServe(*bind, nil)
}

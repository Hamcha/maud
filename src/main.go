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
var adminConf AdminConfig

// absolute path to Maud root directory
var maudRoot string

func setupHandlers(router *mux.Router, isAdmin, isSubdir bool) {
	GET := router.Methods("GET").Subrouter()
	POST := router.Methods("POST").Subrouter()

	SetHandler(GET, "/", httpHome, isAdmin, isSubdir)
	SetHandler(GET, "/tag/{tag}", httpTagSearch, isAdmin, isSubdir)
	SetHandler(GET, "/tag/{tag}/page/{page}", httpTagSearch, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}", httpThread, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}/page/{page}", httpThread, isAdmin, isSubdir)
	SetHandler(GET, "/new", httpNewThread, isAdmin, isSubdir)
	SetHandler(GET, "/threads", httpAllThreads, isAdmin, isSubdir)
	SetHandler(GET, "/threads/page/{page}", httpAllThreads, isAdmin, isSubdir)
	SetHandler(GET, "/tags", httpAllTags, isAdmin, isSubdir)
	SetHandler(GET, "/tags/page/{page}", httpAllTags, isAdmin, isSubdir)
	SetHandler(GET, "/wiki", httpWikiIndex, isAdmin, isSubdir)
	SetHandler(GET, "/wiki/{page}", httpWiki, isAdmin, isSubdir)

	SetHandler(POST, "/new", apiNewThread, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/reply", apiReply, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/post/{post}/edit", apiEditPost, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/post/{post}/delete", apiDeletePost, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/post/{post}/raw", apiGetRaw, isAdmin, isSubdir)
	SetHandler(POST, "/tagsearch", apiTagSearch, isAdmin, isSubdir)
	SetHandler(POST, "/postpreview", apiPreview, isAdmin, isSubdir)
	SetHandler(POST, "/taglist", apiTagList, isAdmin, isSubdir)
}

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
	adminfile := flag.String("admin", "admin.conf", "Admin configuration file")
	flag.StringVar(&maudRoot, "root", maudRoot, "The HTTP server root directory")
	flag.Parse()

	// Load Site info file
	rawconf, _ := ioutil.ReadFile(maudRoot + "/info.json")
	err = json.Unmarshal(rawconf, &siteInfo)
	if err != nil {
		panic(err)
	}

	// Load Admin config file
	if (*adminfile)[0] != '/' {
		*adminfile = maudRoot + "/" + *adminfile
	}
	rawadmin, _ := ioutil.ReadFile(*adminfile)
	err = json.Unmarshal(rawadmin, &adminConf)
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

	// Admin mode pages
	initAdmin()
	if adminConf.EnablePath {
		adminPath := router.PathPrefix(adminConf.Path).Subrouter()
		setupHandlers(adminPath, true, true)
	}
	if adminConf.EnableDomain {
		adminHost := router.Host(adminConf.Domain).Subrouter()
		setupHandlers(adminHost, true, false)
	}

	setupHandlers(router, false, false)
	http.Handle("/", router)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(maudRoot+"/static"))))

	// Start serving pages
	fmt.Printf("Listening on %s\r\nServer root: %s\r\n", *bind, maudRoot)
	http.ListenAndServe(*bind, nil)
}

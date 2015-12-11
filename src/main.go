package main

import (
	. "./data"
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	adminConf AdminConfig
	db        Database
	maudRoot  string
	siteInfo  SiteInfo
)

func setupHandlers(router *mux.Router, isAdmin, isSubdir bool) {
	GET := router.Methods("GET", "HEAD").Subrouter()
	POST := router.Methods("POST").Subrouter()

	SetHandler(GET, "/", httpHome, isAdmin, isSubdir)
	SetHandler(GET, "/tag/{tag}", httpTagSearch, isAdmin, isSubdir)
	SetHandler(GET, "/tag/{tag}/page/{page}", httpTagSearch, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}", httpThread, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}/page/{page}", httpThread, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}/post/{post}/edit", httpEditPost, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}/post/{post}/delete", httpDeletePost, isAdmin, isSubdir)
	SetHandler(GET, "/thread/{thread}/post/{post}/ban", httpBanUser, isAdmin, isSubdir)
	SetHandler(GET, "/new", httpNewThread, isAdmin, isSubdir)
	SetHandler(GET, "/threads", httpAllThreads, isAdmin, isSubdir)
	SetHandler(GET, "/threads/page/{page}", httpAllThreads, isAdmin, isSubdir)
	SetHandler(GET, "/tags", httpAllTags, isAdmin, isSubdir)
	SetHandler(GET, "/tags/page/{page}", httpAllTags, isAdmin, isSubdir)
	SetHandler(GET, "/stiki", httpStikiIndex, isAdmin, isSubdir)
	SetHandler(GET, "/stiki/{page}", httpStiki, isAdmin, isSubdir)
	SetHandler(GET, "/hidden", httpManageHidden, isAdmin, isSubdir)
	SetHandler(GET, "/hidden/page/{page}", httpManageHidden, isAdmin, isSubdir)
	SetHandler(GET, "/blacklist", httpBlacklist, isAdmin, isSubdir)
	SetHandler(GET, "/{otherwise}", func(rw http.ResponseWriter, req *http.Request) {
		sendError(rw, 404, "Not found")
	}, isAdmin, isSubdir)

	SetHandler(POST, "/new", apiNewThread, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/reply", apiReply, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/post/{post}/edit", apiEditPost, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/post/{post}/delete", apiDeletePost, isAdmin, isSubdir)
	SetHandler(POST, "/thread/{thread}/post/{post}/raw", apiGetRaw, isAdmin, isSubdir)
	SetHandler(POST, "/tagsearch", apiTagSearch, isAdmin, isSubdir)
	SetHandler(POST, "/postpreview", apiPreview, isAdmin, isSubdir)
	SetHandler(POST, "/taglist", apiTagList, isAdmin, isSubdir)
	SetHandler(POST, "/blacklist/new", apiBlacklistAdd, isAdmin, isSubdir)
	SetHandler(POST, "/blacklist/{rule}/edit", apiBlacklistEdit, isAdmin, isSubdir)
	SetHandler(POST, "/blacklist/{rule}/delete", apiBlacklistRemove, isAdmin, isSubdir)
	SetHandler(POST, "/blacklist/{rule}/raw", apiBlacklistGetRaw, isAdmin, isSubdir)
}

func dontListDirs(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			sendError(w, 403, "Forbidden")
			return
		}
		h.ServeHTTP(w, r)
	})
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
	err = loadJson("info.json", &siteInfo)
	if err != nil {
		panic(err)
	}

	// Load Admin config file
	err = loadJson(*adminfile, &adminConf)
	if err != nil {
		log.Println("[ WARNING ] Admin file is missing or malformed, Maud will run without administrators.")
	}

	// Initialize formatters, database and other modules
	db = InitDatabase(*mongo, *dbname)
	defer db.Close()
	InitFormatters()

	// Initialize blacklist
	initBL()

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
	http.Handle("/static/", dontListDirs(http.StripPrefix("/static/", http.FileServer(http.Dir(maudRoot+"/static")))))

	// Start serving pages
	log.Printf("Listening on %s\r\nServer root: %s\r\n", *bind, maudRoot)
	http.ListenAndServe(*bind, nil)
}

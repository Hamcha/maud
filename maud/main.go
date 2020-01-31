package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"
)

var (
	db       Database
	maudRoot string
	footers  []string
	csp      = map[string]string{
		"script-src": "'self'",
		"style-src":  "'self'",
		"font-src":   "'self'",
		//"object-src": "'none'",
	}
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
	SetHandler(GET, "/vars.js", httpVars, isAdmin, isSubdir)
	SetHandler(GET, "/robots.txt", httpRobots, isAdmin, isSubdir)
	SetHandler(GET, "/{otherwise}", func(rw http.ResponseWriter, req *http.Request) {
		if isEmoji, surl := emojiRedir(req); isEmoji {
			http.Redirect(rw, req, "/thread/"+surl, http.StatusMovedPermanently)
			return
		}
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

var debugMode bool

func mustGet(key string) string {
	val := viper.GetString(key)
	if val == "" {
		panic(fmt.Errorf("config key %s not set to a value", key))
	}
	return val
}

func main() {
	// Set default values
	viper.SetDefault("bind", ":8080")
	viper.SetDefault("dburl", "localhost")
	viper.SetDefault("dbname", "maud")
	viper.SetDefault("fullDomain", "localhost.localdomain:8080")
	viper.SetDefault("liteDomain", "lite.localhost.localdomain:8080")
	viper.SetDefault("debug", false)
	viper.SetDefault("postsPerPage", 30)
	viper.SetDefault("threadsPerPage", 20)
	viper.SetDefault("tagsPerPage", 20)
	viper.SetDefault("tagResultsPerPage", 10)
	viper.SetDefault("tagsInHome", 5)
	viper.SetDefault("threadsInHome", 10)
	viper.SetDefault("maxPostLength", 15000)
	viper.SetDefault("useProxy", false)
	viper.SetDefault("siteTitle", "Maud pie lair")

	// Setup sources
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("maud")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Could not find config file, maud will use other sources if possible (or fail)")
		} else {
			panic(err)
		}
	}

	bind := viper.GetString("bind")
	debugMode = viper.GetBool("debug")
	maudRoot = viper.GetString("maudroot")
	if maudRoot == "" {
		maudRoot, _ = os.Getwd()
	}

	// Read footers

	if footersTxt, err := ioutil.ReadFile(maudRoot + "/footers.txt"); err == nil {
		footers = strings.Split(string(footersTxt), "\n")
	} else {
		footers = []string{}
	}

	// Initialize formatters, database and other modules
	mongo := viper.GetString("dburl")
	dbname := viper.GetString("dbname")
	log.Printf("[ INFO ] Connecting to %s/%s ...", mongo, dbname)
	db = InitDatabase(mongo, dbname)
	defer db.Close()
	log.Printf("[ OK ] Connected.\r\n")
	InitFormatters()

	// Initialize blacklist
	initBL()

	// Setup request handlers
	router := mux.NewRouter()

	// Admin mode pages
	initAdmin()
	if viper.GetBool("adminEnablePath") {
		adminPath := router.PathPrefix(mustGet("adminPath")).Subrouter()
		setupHandlers(adminPath, true, true)
	}
	if viper.GetBool("adminEnableDomain") {
		adminHost := router.Host(mustGet("adminDomain")).Subrouter()
		setupHandlers(adminHost, true, false)
	}

	setupHandlers(router, false, false)
	http.Handle("/", router)
	http.Handle("/static/", dontListDirs(http.StripPrefix("/static/", http.FileServer(http.Dir(maudRoot+"/static")))))

	// Start serving pages
	log.Printf("Listening on %s\r\nServer root: %s\r\n", bind, maudRoot)
	http.ListenAndServe(bind, nil)
}

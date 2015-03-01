package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

type Blacklist struct {
	Criteria  string
	IP        string
	UserAgent string
	Reason    string

	IPRegexp *regexp.Regexp
	UARegexp *regexp.Regexp
}

var blacklists map[string]Blacklist

func initBL() {
	// Load Site info file
	rawconf, err := ioutil.ReadFile(maudRoot + "/blacklist.conf")
	if err != nil {
		log.Printf("[ WARNING ] %s/blacklist.conf not found. BL not initialized.\n", maudRoot)
		return
	}
	err = json.Unmarshal(rawconf, &blacklists)
	if err != nil {
		panic(err)
	}
	for i := range blacklists {
		item := blacklists[i]
		item.IPRegexp = regexp.MustCompile(blacklists[i].IP)
		item.UARegexp = regexp.MustCompile(blacklists[i].UserAgent)
		blacklists[i] = item
	}
}

func checkBlacklist(req *http.Request) (bool, string) {
	userAgent := req.UserAgent()
	ip := req.RemoteAddr
	for i := range blacklists {
		ipmatch := blacklists[i].IPRegexp.MatchString(ip)
		uamatch := blacklists[i].UARegexp.MatchString(userAgent)
		var matches bool
		if blacklists[i].Criteria == "ALL" {
			matches = ipmatch && uamatch
		} else {
			matches = ipmatch || uamatch
		}

		if matches {
			return true, blacklists[i].Reason
		}
	}

	return false, ""
}

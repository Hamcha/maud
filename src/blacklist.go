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
	Action    string // "ban": shows 423, "captcha": asks for captcha on reply/new thread

	IPRegexp *regexp.Regexp
	UARegexp *regexp.Regexp
}

var blacklists map[string]Blacklist
var captchas []CaptchaData

func initBL() {
	// Load Blacklist conf
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
	// Load Captcha conf
	rawconf, err = ioutil.ReadFile(maudRoot + "/captcha.conf")
	if err != nil {
		log.Printf("[ WARNING %s/captcha.conf not found. Captchas not initialized.\n", maudRoot)
		return
	}
	err = json.Unmarshal(rawconf, &captchas)
	if err != nil {
		panic(err)
	}
}

// checkBlacklists finds out if a request comes from a blacklisted IP.
// Returns (isBanned, banReason, banAction).
func checkBlacklist(req *http.Request) (bool, string, string) {
	userAgent := req.UserAgent()
	var ip string
	if iphead, ok := req.Header["X-Forwarded-For"]; ok {
		ip = iphead[0]
	} else {
		ip = req.RemoteAddr
	}
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
			return true, blacklists[i].Reason, blacklists[i].Action
		}
	}

	return false, "", ""
}

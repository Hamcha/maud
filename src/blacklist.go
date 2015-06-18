package main

import (
	"./data"
	"github.com/oschwald/maxminddb-golang"
	"log"
	"net"
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
var captchas []data.CaptchaData
var geoip *maxminddb.Reader

type onlyCountry struct {
	Country struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func initBL() {
	// Load Blacklist conf
	err := LoadJson("blacklist.conf", &blacklists)
	if err != nil {
		log.Printf("[ WARNING ] %s/blacklist.conf not found. BL not initialized.\n", maudRoot)
	} else {
		for i := range blacklists {
			item := blacklists[i]
			item.IPRegexp = regexp.MustCompile(blacklists[i].IP)
			item.UARegexp = regexp.MustCompile(blacklists[i].UserAgent)
			blacklists[i] = item
		}
	}

	// Load Captcha conf
	err = LoadJson("captcha.conf", &captchas)
	if err != nil {
		log.Printf("[ WARNING ] %s/captcha.conf not found. Captchas not initialized.\n", maudRoot)
	}

	// Load GeoIP database
	geoip, err = maxminddb.Open("geoip.mmdb")
	if err != nil {
		log.Printf("[ WARNING ] %s/geoip.mmdb not found. GeoIP limiting not initialized.\n", maudRoot)
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

	// Check on blacklist.conf
	if blacklists != nil {
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
	}

	// Check on GeoIP database
	if geoip != nil {
		ip := net.ParseIP(ip)
		var result onlyCountry
		geoip.Lookup(ip, &result)

		if result.Country.IsoCode != "IT" {
			return true, "", "captcha"
		}
	}

	return false, "", ""
}

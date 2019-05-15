package main

import (
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"

	. "github.com/hamcha/maud/maud/data"
	"github.com/oschwald/maxminddb-golang"
)

type BlacklistParams struct {
	Criteria  string
	IP        string
	UserAgent string
	Reason    string
	Action    string
}

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
var geoip *maxminddb.Reader

type onlyCountry struct {
	Country struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

func initBL() {
	// Load Blacklist conf
	err := loadJson(confRoot, "blacklist.conf", &blacklists)
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
	err = loadJson(confRoot, "captcha.conf", &captchas)
	if err != nil {
		log.Printf("[ WARNING ] %s/captcha.conf not found. Captchas not initialized.\n", maudRoot)
	}

	// Load GeoIP database
	geoip, err = maxminddb.Open(maudRoot + "/geoip.mmdb")
	if err != nil {
		log.Printf("[ WARNING ] %s/geoip.mmdb not found. GeoIP limiting not initialized.\n", maudRoot)
	}
}

func wrapBlacklist(handler http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		isBanned, banReason, blAction := checkBlacklist(req)
		if isBanned {
			switch blAction {
			case "ban":
				sendError(rw, 423, banReason)
				return
			case "captcha":
				req.Header.Add("Captcha-required", "true")
			}
		}
		handler(rw, req)
	}
}

// checkBlacklists finds out if a request comes from a blacklisted IP.
// Returns (isBanned, banReason, banAction).
func checkBlacklist(req *http.Request) (bool, string, string) {
	if a, _ := isAdmin(req); a {
		return false, "", ""
	}
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
			if strings.ToLower(blacklists[i].Criteria) == "all" {
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

// Parameters returns a struct with only the conf params (i.e.
// without IPRegexp and UARegexp)
func (blacklist Blacklist) Parameters() (params BlacklistParams) {
	params.Criteria = blacklist.Criteria
	params.IP = blacklist.IP
	params.UserAgent = blacklist.UserAgent
	params.Reason = blacklist.Reason
	params.Action = blacklist.Action
	return
}

func NewBlacklist(params BlacklistParams) Blacklist {
	return Blacklist{
		Criteria:  params.Criteria,
		IP:        params.IP,
		UserAgent: params.UserAgent,
		Reason:    params.Reason,
		Action:    params.Action,
		IPRegexp:  regexp.MustCompile(params.IP),
		UARegexp:  regexp.MustCompile(params.UserAgent),
	}
}

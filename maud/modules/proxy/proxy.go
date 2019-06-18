package proxy

import (
	"crypto/tls"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hamcha/maud/maud/modules"
)

func Provide(root, domain string) modules.PostMutator {
	pm := new(proxyMutator)
	pm.init(root, domain)
	return modules.PostMutator{
		Condition: isProxyEnabled,
		Mutator:   pm.mutator,
	}
}

type proxyMutator struct {
	imgRgx *regexp.Regexp
	proxy  Proxy
	domain string
}

func (f *proxyMutator) init(root, domain string) {
	f.imgRgx = regexp.MustCompile(`<img [^>]*src=["']?([^"']+)["']?[^>]*>`)
	f.proxy.Root = root
	f.proxy.MaxWidth = 640
	f.proxy.MaxHeight = 300

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	f.proxy.client = http.Client{
		Timeout:   time.Duration(10 * time.Second),
		Transport: transport,
	}
	f.domain = domain
	log.Printf("[ OK ] Module initialized: Proxy (root: " + root + ", domain: " + domain + ")")
}

func isProxyEnabled(req *http.Request) bool {
	_, err := req.Cookie("crUseProxy")
	return err == nil
}

// Mutator converts all external <img> references to the relative
// cached URLs (as configured in ProxyDomain). If the resource is
// not currently cached, caches it. If fetching fails somehow,
// the <img> is replaced with a simple link to the original resource.
func (f *proxyMutator) mutator(data modules.PostMutatorData) {
	rawcontent := data.Post.Content
	type Content struct {
		Original string // The original HTML tag
		URL      string // The URL of the external image
	}
	imgChans := make(map[Content]<-chan AsyncImageResult)
	matches := f.imgRgx.FindAllStringSubmatch(rawcontent, -1)
	for _, match := range matches {
		content := Content{
			Original: match[0],
			URL:      match[1],
		}
		imgChans[content] = f.proxy.GetImageAsync(content.URL)
	}
	for content, uchan := range imgChans {
		res := <-uchan
		if res.Error == nil {
			sizedata := ""
			if res.Data.Width != 0 || res.Data.Height != 0 {
				sizedata = ` width="` + strconv.Itoa(res.Data.Width) +
					`" height="` + strconv.Itoa(res.Data.Height) + `"`
			}
			// Serve the cached content
			rawcontent = strings.Replace(rawcontent, content.Original,
				`<img src="`+f.domain+res.Path+`"`+sizedata+` alt="`+content.URL+`">`, -1)
		} else {
			// Give up and serve the link instead
			rawcontent = strings.Replace(rawcontent, content.Original,
				`<a href="`+content.URL+`">`+content.URL+`</a>`, -1)
		}
	}
	(*data.Post).Content = rawcontent

	// Force 'img-src: self' in CSP
	f.addImgSrcCSPRule(data.ResponseWriter)
}

func (f *proxyMutator) addImgSrcCSPRule(rw *http.ResponseWriter) {
	head := (*rw).Header()
	if !strings.Contains(head.Get("Content-Security-Policy"), "img-src") {
		head.Add("Content-Security-Policy", "img-src 'self' "+f.domain)
	}
}

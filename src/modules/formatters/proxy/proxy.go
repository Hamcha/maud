package proxy

import (
	"log"
	"regexp"
	"strings"
)

func Provide(root, domain string) *ProxyFormatter {
	proxyFormatter := new(ProxyFormatter)
	proxyFormatter.init(root, domain)
	return proxyFormatter
}

type ProxyFormatter struct {
	imgRgx *regexp.Regexp
	proxy  Proxy
	domain string
}

func (f *ProxyFormatter) init(root, domain string) {
	f.imgRgx = regexp.MustCompile(`<img [^>]*src=["']?([^"']+)["']?[^>]*>`)
	f.proxy.Root = root
	f.domain = domain
	log.Printf("Module initialized: Proxy")
}

// Mutator converts all external <img> references to the relative
// cached URLs (as configured in ProxyDomain). If the resource is
// not currently cached, caches it. If fetching fails somehow,
// the <img> is replaced with a simple link to the original resource.
func (f *ProxyFormatter) Format(rawcontent string) string {
	type Content struct {
		Type     string
		Original string
		URL      string
	}
	chans := make(map[Content]<-chan string)
	matches := f.imgRgx.FindAllStringSubmatch(rawcontent, -1)
	for _, match := range matches {
		content := Content{
			Type:     "image",
			Original: match[0],
			URL:      match[1],
		}
		chans[content] = f.proxy.GetContentAsync(content.URL)
	}
	for content, uchan := range chans {
		proxyUrl := <-uchan
		if proxyUrl != "" {
			// Serve the cached content
			rawcontent = strings.Replace(rawcontent, content.Original,
				`<img src="`+f.domain+proxyUrl+`" alt="`+content.URL+`">`, -1)
		} else {
			// Give up and serve the link instead
			rawcontent = strings.Replace(rawcontent, content.Original, `<a href="`+content.URL+`">`+content.URL+`</a>`, -1)
		}
	}
	return rawcontent
}

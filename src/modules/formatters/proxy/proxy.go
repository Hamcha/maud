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
func (f *ProxyFormatter) Format(content string) string {
	for _, match := range f.imgRgx.FindAllStringSubmatch(content, -1) {
		origUrl := match[1]
		if proxyUrl, err := f.proxy.GetContent(origUrl); err == nil {
			// Serve the cached content
			content = strings.Replace(content, match[0],
				`<img src="`+f.domain+proxyUrl+`" alt="`+origUrl+`">`, -1)
		} else {
			// Give up and serve the link instead
			content = strings.Replace(content, match[0], `<a href="`+origUrl+`">`+origUrl+`</a>`, -1)
		}
	}
	return content
}

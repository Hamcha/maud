package proxy

import (
	".."
	"../.."
	"net/http"
	"os"
	"path"
	"strings"
)

func Provide() *ProxyFormatter {
	proxyFormatter := new(ProxyFormatter)
	proxyFormatter.Init()
	return proxyFormatter
}

type ProxyFormatter struct {
	imgRgx *regexp.Regexp
	proxy  Proxy
}

func (f *ProxyFormatter) Init(root string) {
	f.imgRgx = regexp.MustCompile(`(?:<a [^>]+>)?<img .*src=("[^"]+"|'[^']+'|[^'"][^\s]+).*>(?:</a>)?`)
	f.proxy.Root = root
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
			content = strings.Replace(content, match[0], modules.WrapImg(origUrl, proxyUrl, nil), -1)
		} else {
			// Give up and serve the link instead
			content = strings.Replace(content, match[0], `<a href="`+origUrl+`">`+origUrl+`</a>`)
		}
	}
	return content
}

package proxy

import (
	".."
	"../.."
	"net/http"
	"os"
	"path"
	"strings"
)

func Provide() modules.PostMutator {
	proxyMutator := new(ProxyMutator)
	proxyMutator.Init()
	return proxyMutator
}

type ProxyMutator struct {
	imgRgx *regexp.Regexp
}

func (f *ProxyMutator) Init() {
	f.imgRgx = regexp.MustCompile(`(?:<a [^>]+>)?<img .*src=("[^"]+"|'[^']+'|[^'"][^\s]+).*>(?:</a>)?`)
}

func (f *ProxyMutator) Condition(_ *http.Request) bool {
	return siteInfo.UseProxy
}

// Mutator converts all external <img> references to the relative
// cached URLs (as configured in ProxyDomain). If the resource is
// not currently cached, caches it. If fetching fails somehow,
// the <img> is replaced with a simple link to the original resource.
func (f *ProxyMutator) Mutator(content string) string {
	for _, match := range f.imgRgx.FindAllStringSubmatch(content, -1) {
		origUrl := match[1]
		if proxyUrl, err := someproxyhere.GetContent(origUrl); err == nil {
			// Serve the cached content
			content = strings.Replace(content, match[0], modules.WrapImg(origUrl, proxyUrl, nil), -1)
		} else {
			// Give up and serve the link instead
			content = strings.Replace(content, match[0], `<a href="`+origUrl+`">`+origUrl+`</a>`)
		}
	}
	return content
}

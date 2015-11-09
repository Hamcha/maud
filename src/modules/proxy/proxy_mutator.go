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
		cachedUrl := getProxiedUrl(origUrl)
		if _, err := os.Stat(path.Join(siteInfo.ProxyRoot)); err == nil {
			content = strings.Replace(content, match[0], WrapImg(origUrl, cachedUrl, nil), -1)
		}
	}
	return content
}

// getProxydUrl replaces an URL like http://domain.com/path/to/file to
// https://<siteInfo.ProxyDomain>/domain.com/path/to/file
func getProxiedUrl(orig string) string {
	split := strings.Split(orig, '/')
	split[0] = "https"
	split[3] = split[2] + split[3]
	split[2] = siteInfo.ProxyDomain
	return strings.Join(split, '/')
}

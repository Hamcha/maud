package proxy

import (
	"../.."
	"log"
	"net/http"
	"regexp"
	"strings"
)

func Provide(root, domain string) modules.PostMutator {
	pm := new(proxyMutator)
	pm.init(root, domain)
	return modules.PostMutator{
		Condition: condition,
		Mutator: pm.mutator,
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
	f.domain = domain
	log.Printf("Module initialized: Proxy")
}

func condition(req *http.Request) bool {
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
	(*data.Post).Content = rawcontent
}

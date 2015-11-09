package proxy

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type Proxy struct {
	Root string
}

// GetContent retrieves a content from a remote server and returns
// the relative path to it
// If the content already exists then it just returns the path to it
func (p Proxy) GetContent(contentURL string) (string, error) {
	contentPath := getPathURL(contentURL)
	ospath := Proxy.Root + filepath.FromSlash(contentPath)

	_, err := os.Stat(ospath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, fetch it
			err = Fetch(contentURL)
			if err != nil {
				return contentPath, err
			}
		} else {
			return contentPath, err
		}
	}

	// File exists or has been fetched, return path to it
	return contentPath, nil
}

func (p Proxy) Fetch(contentURL string) error {
	resp, err := http.Get(contentURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	path := getPathURL(contentURL)
	ospath := p.Root + filepath.FromSlash(path)

	// Create the directory tree
	err = os.MkdirAll(ospath, 0600)
	if err != nil {
		return err
	}

	// Create the file
	file, err := os.Create(ospath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	return err
}

func getPathURL(rawURL string) string {
	urldata := url.Parse(rawURL)
	return "/" + urldata.Host + urldata.Path
}

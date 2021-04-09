
package rest

import (
	"fmt"
	"net/http"
	"io/ioutil"	
)

type Client interface {
	Get(url string) (*http.Response, error)
	GetBody(url string) (string, error)
}

type HttpClient struct {
}

// something like this? https://freshman.tech/snippets/go/http-response-to-string/
func (client *HttpClient) GetBody(url string) (string, error) {
	resp, geterr := http.Get(url)
	if geterr != nil {
		return "", geterr
	}

    if resp.StatusCode < 200 || resp.StatusCode > 299 {
    	return "", fmt.Errorf("Status code for %s is not OK (%d)", url, resp.StatusCode)
    }

	defer resp.Body.Close()
	body, readerr := ioutil.ReadAll(resp.Body)
	if readerr != nil {
		return "", readerr
	}

	return string(body), nil
}


func (client *HttpClient) Get(url string) (*http.Response, error) {
	return http.Get(url)
}
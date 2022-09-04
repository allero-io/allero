package httpClient

import (
	"io/ioutil"
	"net/http"
)

type HttpClient struct {
	baseUrl string
}

func New() (*HttpClient, error) {
	return &HttpClient{
		baseUrl: "https://api.allero.io",
	}, nil
}

func (c *HttpClient) Get(relativeUrl string) ([]byte, error) {
	var url string
	if relativeUrl == "" {
		url = c.baseUrl
	} else {
		url = c.baseUrl + "/" + relativeUrl
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

package robospider

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type httpClient struct {
	proxyServer string
}


// Ensure the domain has the protocol
func BuildDomainURL(input string) string {
	match, _ := regexp.MatchString("^(https?://)", input)
	if match == false {
		input = fmt.Sprintf("http://%v", input)
	}
	return input
}

func (cli *httpClient) Fetch(targetUrl *url.URL, result chan<- Resource) error {
	hc := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	if cli.proxyServer != "" {
		// Parse the proxy address
		log.Println("[d]: Trying to validate the proxy address before using it.")
		parsedURL, parseErr := url.Parse(BuildDomainURL(cli.proxyServer))

		// Warn the proxy error and stop the execution to prevent any unwanted request
		if parseErr != nil {
			log.Println("[e]: Invalid proxy address:", parseErr)
			return parseErr
		}

		// Set the http client proxy and increase default timeout since proxy can be slow as fuck
		log.Println("[i]: Setting up transport with proxy server at address:", parsedURL)
		hc.Transport = &http.Transport{Proxy: http.ProxyURL(parsedURL)}
		hc.Timeout = time.Duration(10 * time.Second)
	}

	resp, err := hc.Get(targetUrl.String())

	// if request failed show the error and exit
	if err != nil && resp.StatusCode != http.StatusNotFound {
		fmt.Println("[e]: Failed to fetch resource:", err)
		return err
	}

	// DO NOT close the response, as the body is going to be put in a channel
	// the caller must close it after use
	//defer resp.Body.Close()

	result <- Resource{
		Name:  targetUrl.String(),
		Found: (resp.StatusCode == http.StatusOK),
		Body:  resp.Body,
	}

	return nil
}

func NewHttpClient() *httpClient {
	return &httpClient{
		proxyServer: "",
	}
}

func NewHttpClientWithProxy(proxyURL string) *httpClient {
	return &httpClient{
		proxyServer: proxyURL,
	}
}


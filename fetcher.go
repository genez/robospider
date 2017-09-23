package robospider

import "net/url"

type Fetcher interface {
	Fetch(targetURL *url.URL, result chan<- Resource) error
}

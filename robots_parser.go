package robospider

import (
	"bufio"
	"io"
	"net/url"
	"strings"
	"log"
)

type robotsParser struct {
	robotsURL *url.URL
}

func NewRobotsParser(robotsURL *url.URL) *robotsParser {
	return &robotsParser{robotsURL}
}

// Read text line by line and get robot file entries
func (r *robotsParser) Parse(input io.Reader) ([]*url.URL, error) {
	var entries []*url.URL
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		value := strings.Split(scanner.Text(), "Disallow: /")
		if len(value) > 1 {
			relative, err := url.Parse(value[1])
			if err != nil {
				return entries, err
			}
			resolved := r.robotsURL.ResolveReference(relative)
			log.Println("[i] resolved URL: ", resolved.String())
			entries = append(entries, resolved)
		}
	}
	return entries, scanner.Err()
}

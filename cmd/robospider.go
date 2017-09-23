package main

import (
	"flag"
	"fmt"
	"github.com/genez/robospider"
	"io"
	"log"
	"net/url"
	"os"
	"sync"
	"time"
)

const version = "1.0.0"

var proxy = flag.String("proxy", "", "the full address of the proxy server to use: [address:port]")
var output = flag.String("output", "", "the output file name Default: [domain].log")

// Usage print a custom usage function
func Usage() {
	fmt.Print("Package usage: robospider [-proxy URL] [-output NAME] [DOMAIN]\n\n")
	flag.PrintDefaults()
}

// Print the project banner
func printBanner() {
	fmt.Println("          ")
	fmt.Println("   |  |   ")
	fmt.Printf("   \\**/    Robospider v%v \n", version)
	fmt.Println("  o={}=o   by Filippo 'b4dnewz' Conti")
	fmt.Println(" / /()\\ \\  codekraft-studio <info@codekraft.it>")
	fmt.Println("   \\  /  ")
	fmt.Println("          ")
}

const workerPoolSize = 8

func main() {
	flag.Usage = Usage

	// Output package banner
	printBanner()

	// get script arguments
	args := os.Args[1:]

	// Exit with the usage informations if no arguments
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Parse script flags into variables
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(1)
	}

	// robots file is case sensitive and must be placed in the root directory
	// so this url construction pattern should always match
	robotsURL := fmt.Sprintf("%s/robots.txt", args[0])
	// Parse url string into object
	u, err := url.ParseRequestURI(robospider.BuildDomainURL(robotsURL))
	// Exit if url can't be parsed
	if err != nil {
		log.Fatal("[e]: Invalid site address:", robotsURL, err)
	}

	//buffer size should be tuned with real world tests
	robotsResources := make(chan robospider.Resource, 1)

	client := robospider.NewHttpClientWithProxy(*proxy)
	err = client.Fetch(u, robotsResources)
	if err != nil {
		log.Fatal("[e]: Robots.txt download error:", err)
	}

	var robotsTxt robospider.Resource
	select {
	case robotsTxt = <-robotsResources:
		defer robotsTxt.Body.Close()

	case <-time.After(5 * time.Second):
		log.Fatal("[e]: Robots.txt download timeout:", err)
	}

	rp := robospider.NewRobotsParser(u)
	disallowedEntries, err := rp.Parse(robotsTxt.Body)
	if err != nil {
		log.Fatal("[e]: Robots.txt parsing failed:", err)
	}

	// Create the output directory
	err = os.Mkdir("output", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal("[e]: could not create output directory:", err)
	}

	// Set default output name if nothing was passed
	if *output == "" {
		*output = fmt.Sprintf("output/%v.log", args[0])
	} else {
		*output = fmt.Sprintf("output/%v.log", output)
	}

	//we have the stream of robots.txt file now
	//let's create a channel for subsequent workers
	//buffered channel size should be tuned according to benchmarks/real world tests

	//buffer size should be tuned with real world tests
	disallowedResources := make(chan robospider.Resource, workerPoolSize)

	start := time.Now()

	wgProducers := &sync.WaitGroup{}
	//let's start a new goroutine for each downloader (producer)
	for i := 0; i < workerPoolSize && i < len(disallowedEntries); i++ {
		wgProducers.Add(1)
		go func(targetURL *url.URL) {
			defer wgProducers.Done()

			c := robospider.NewHttpClientWithProxy(*proxy)
			err := c.Fetch(targetURL, disallowedResources)
			if err != nil {
				log.Fatal("[e]:", targetURL, ": download error:", err)
			}
		}(disallowedEntries[i])
	}

	successCount := 0

	wgConsumers := &sync.WaitGroup{}
	//let's start consumers on the result channel (in a new goroutine)
	for i := 0; i < workerPoolSize && i < len(disallowedEntries); i++ {
		wgConsumers.Add(1)
		go func() {
			defer wgConsumers.Done()

			for r := range disallowedResources {
				if r.Body != nil && r.Found {
					writeFile(r.Body, r.Name)
					r.Body.Close()
					successCount++
				}
			}
		}()
	}

	//wait all the downloaders to complete their work
	wgProducers.Wait()
	//closing the channel will tell consumers there are no more results to handle
	close(disallowedResources)
	//wait for all consumers to finish their work
	wgConsumers.Wait()

	// Output the scan time
	fmt.Printf("\n[i]: The scan has completed with %v error and %v success in %v.\n", len(disallowedEntries)-successCount, successCount, time.Since(start))
}

func writeFile(reader io.Reader, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}
	return nil
}

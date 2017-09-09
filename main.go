package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const version = "1.0.0"

var proxy string
var output string
var successList []string

// Init sscript flag variables
func init() {

	flag.StringVar(&output, "output", "", "the output file name Default: [domain].log")
	flag.StringVar(&proxy, "proxy", "", "the full address of the proxy server to use: [address:port]")

}

// Print the project banner
func printBanner() {

	fmt.Println("          ")
	fmt.Println("   |  |   ")
	fmt.Printf("   \\()/    Robospider v%v \n", version)
	fmt.Println("  o={}=o   by Filippo 'b4dnewz' Conti")
	fmt.Println(" / /**\\ \\  codekraft-studio <info@codekraft.it>")
	fmt.Println("   \\  /  ")
	fmt.Println("          ")

}

// Validate input as url
func validateDomain(input string) (bool, error) {
	return regexp.MatchString("^(https?://)?([\\da-z\\.-]+)\\.([a-z\\.]{2,6})$", input)
}

// Extract the domain name from the url
func getDomainName(input string) string {
	r := regexp.MustCompile("(?i)^(?:https?://)?(?:[^@\n]+@)?(?:www\\.)?([^:/\n]+)")
	return r.FindStringSubmatch(input)[1]
}

// Ensure the domain has the protocol
func buildDomainURL(input string) string {

	match, _ := regexp.MatchString("^(https?://)", input)

	if match == false {
		input = fmt.Sprintf("http://%v", input)
	}

	return input

}

// Write resulting lines to file
func writeLines(lines []string, path string) error {

	file, err := os.Create(path)

	if err != nil {
		return err
	}

	defer file.Close()

	w := bufio.NewWriter(file)

	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	return w.Flush()

}

// Read text line by line and get robot file entries
func readLines(input string) ([]string, error) {

	var lines []string

	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		value := strings.Split(scanner.Text(), "Disallow: /")
		if len(value) > 1 {
			lines = append(lines, value[1])
		}
	}

	return lines, scanner.Err()

}

func main() {

	// Output package banner
	printBanner()

	// get script arguments
	args := os.Args[1:]

	// Exit with the usage informations if no arguments
	if len(args) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// TODO: See if this can be done with url.Parse in a more efficient way
	// Exit with an user warning if the domain is not valid url
	if result, _ := validateDomain(args[len(args)-1]); result == false {
		fmt.Println("Warning: You must provide a valid domain, otherwise it will not work.")
		os.Exit(0)
	}

	// Parse script flags into variables
	flag.Parse()

	// client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(parsedURL)}}
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	// If the oroxy flag is set and not empty
	if proxy != "" {

		// Parse the proxy address
		fmt.Println("[d]: Trying to validate the proxy address before using it.")
		parsedURL, err := url.Parse(proxy)

		if err != nil {
			fmt.Println("[e]:", err)
			return
		}

		// Set the http client proxy
		fmt.Println("[i]: Setting up transport with proxy server at address:", parsedURL)
		client.Transport = &http.Transport{Proxy: http.ProxyURL(parsedURL)}

	}

	// TODO: Make an option or something to let user choose from http, https
	domainURL := buildDomainURL(args[len(args)-1])

	// robots file is case sensitive and must be placed in the root directory
	// so this url construction pattern should always match
	robotsURL := fmt.Sprintf("%v/robots.txt", domainURL)

	// try to get the site robot file
	fmt.Println("[i]: Starting scan for domain:", domainURL)
	fmt.Println("[i]: Attempt to get the robot file at address:", robotsURL)
	resp, err := client.Get(robotsURL)

	// if request failed show the error and exit
	if err != nil {
		fmt.Println("[e]:", err)
		os.Exit(0)
	}

	// close response
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// Exit with error if robot file can't be parsed
	if err != nil {
		fmt.Println("[e]: Something went wrong when trying to parse the response:", err)
		os.Exit(0)
	}

	// parse lines into array
	responseString := string(body)
	results, err := readLines(responseString)

	// exit if something when wrong when parsing response
	if err != nil {
		fmt.Println("[e]: Something went wrong when parsing the response:")
		fmt.Println(err)
		os.Exit(0)
	}

	// exit if robot file has no entries
	if len(results) == 0 {
		fmt.Println("[e]: The file doesn't contain any entries to scan, quitting.")
		os.Exit(0)
	}

	// Init execution time counter
	fmt.Printf("[i]: Starting the scan of %v entries:\n\n", len(results))
	start := time.Now()

	// scan each resulting path
	for _, result := range results {

		// create url to check
		pathURL := fmt.Sprintf("%v/%v", domainURL, result)

		fmt.Printf("- Attempt to get result: %v ", pathURL)

		// try to get the url
		resp, err := client.Get(pathURL)

		fmt.Printf("[%v] - %v \n", resp.StatusCode, resp.Status)

		// if the path is not attainable skip to next
		if err != nil || resp.StatusCode != 200 {
			continue
		}

		// add good result to success list
		successList = append(successList, pathURL)

	}

	// Create the output directory
	_ = os.Mkdir("output", os.ModePerm)

	// Set default output name if nothing was passed
	if output == "" {
		output = fmt.Sprintf("output/%v.log", getDomainName(domainURL))
	} else {
		output = fmt.Sprintf("output/%v.log", output)
	}

	// Write the result into output folder
	if err := writeLines(successList, output); err != nil {
		log.Fatalln("[e]:", err)
	}

	fmt.Println("")
	fmt.Printf("[i]: The scan has finished with %v error urls and %v success urls in %v. \n", len(results)-len(successList), len(successList), time.Since(start))
	fmt.Println("[i]: The result file has been created in:", output)

}

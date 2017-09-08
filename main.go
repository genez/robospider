package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const version = "1.0.0"

// TODO: Enable a proxy mode for behind the firewall people

// Build robot full url based on domain
// Try to get the robot file
// Parse it selecting only good results
// For each result check if is spiderable
// Count results and execution time
// Output nice graphic with results

func printBanner() {
	fmt.Println(" ")
	fmt.Println("   |  |  ")
	fmt.Printf("   \\()/    Robospider v%v \n", version)
	fmt.Println("  o={}=o   by Filippo 'b4dnewz' Conti")
	fmt.Println(" / /**\\ \\  codekraft-studio <info@codekraft.it>")
	fmt.Println("   \\  /  ")
	fmt.Println("  ")
}

// read text line by line and get robot file entries
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

	// exit with alert message if no argument provided
	if len(args) == 0 {
		fmt.Println("Warning: You must provide the domain to scan, otherwise it will not work.")
		os.Exit(0)
	}

	// get the domain to scan from arguments
	domain := args[0]

	// TODO: check if it's already a valid url and doesn't need to be constructed
	// TODO: Make an option or something to let user choose from http, https
	domainURL := []string{"http://", domain}

	fmt.Println("[i] Starting scan for domain:", strings.Join(domainURL, ""))

	// robots file is case sensitive and must be placed in the root directory
	// so this url construction pattern should always match
	robotsURL := strings.Join(append(domainURL, "/robots.txt"), "")

	fmt.Println("[i] Trying to get the robots file at url:", robotsURL)

	// try to get the site robot file
	resp, err := http.Get(robotsURL)

	// if request failed show the error and exit
	if err != nil {
		fmt.Println("[e]:", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("[e]: Something went wrong when trying to parse the response:", err)
		os.Exit(1)
	}

	responseString := string(body)

	// parse lines into array
	results, err := readLines(responseString)

	// exit if something when wrong when parsing response
	if err != nil {
		fmt.Println("[e]: Something went wrong when parsing the response:")
		fmt.Println(err)
		os.Exit(1)
	}

	// exit if robot file has no entries
	if len(results) == 0 {
		fmt.Println("[e]: The file doesn't contain any entries to scan, quitting.")
		os.Exit(1)
	}

	// // the list of found urls
	var successList []string

	// Init execution time counter
	start := time.Now()
	fmt.Printf("[i] Starting the scan of %v entries:\n\n", len(results))

	// scan each resulting path
	for _, result := range results {

		// create url to check
		pathURL := strings.Join(append(domainURL, "/", result), "")

		fmt.Printf("- Attempt to get result: %v ", pathURL)

		// try to get the url
		resp, err := http.Get(pathURL)

		// if the path is not attainable skip to next
		if err != nil || resp.StatusCode != 200 {
			fmt.Println("[BAD]")
			continue
		}

		// add good result to success list
		successList = append(successList, pathURL)
		fmt.Println("[OK]")

	}

	// TODO: output success results into file in various format (by flag option)

	fmt.Println("")
	fmt.Printf("[i] The scan has finished with %v error urls and %v success urls. \n", len(results)-len(successList), len(successList))
	fmt.Println("[i] You can find the success results in the output file in your current directory.")
	fmt.Printf("[i] The script took %v to complete. \n\n", time.Since(start))

}

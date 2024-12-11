package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type NuclearCodes struct {
	Alpha       string `json:"alpha"`
	Bravo       string `json:"bravo"`
	Charlie     string `json:"charlie"`
	ValidFrom   string `json:"valid_from"`
	ValidTo     string `json:"valid_to"`
	LastUpdated string `json:"last_updated"`
}

func main() {
	// URL of the website to scrape
	url := "https://www.falloutbuilds.com/fo76/nuke-codes/"

	// Create a custom HTTP client with a User-Agent header
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set a User-Agent header (mimic a real browser)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// Check if the response status is OK
	if resp.StatusCode != 200 {
		log.Fatalf("Error: Status code %d", resp.StatusCode)
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize a struct to store nuke codes and valid time
	codes := &NuclearCodes{}

	// Extract the nuke codes (ALPHA, BRAVO, CHARLIE)
	doc.Find(".d-flex.flex-column.flex-lg-row.justify-content-lg-around.text-center.h3.mb-0").Each(func(i int, s *goquery.Selection) {
		s.Find(".text-nowrap").Each(func(i int, codeSelection *goquery.Selection) {
			// Get the text (nuke codes) inside the <div>
			codeText := codeSelection.Text()

			// Clean up the text (remove extra spaces and labels)
			codeText = strings.TrimSpace(codeText)
			if strings.Contains(codeText, "ALPHA") {
				codes.Alpha = strings.ReplaceAll(codeText, "ALPHA", "")
			}
			if strings.Contains(codeText, "BRAVO") {
				codes.Bravo = strings.ReplaceAll(codeText, "BRAVO", "")
			}
			if strings.Contains(codeText, "CHARLIE") {
				codes.Charlie = strings.ReplaceAll(codeText, "CHARLIE", "")
			}
		})
	})

	// Extract the valid time (from the 'Valid from' and 'to' part)
	doc.Find(".small.mb-4").Each(func(i int, s *goquery.Selection) {
		// Extract "Valid from" and "Valid to" by finding <strong> elements inside
		s.Find("strong").Each(func(i int, strongSelection *goquery.Selection) {
			// Get the text inside the <strong> element
			strongText := strings.TrimSpace(strongSelection.Text())

			// Check for valid time range: "Valid from" and "Valid to"
			if i == 0 {
				codes.ValidFrom = strongText
			} else if i == 1 {
				codes.ValidTo = strongText
			}
		})

		// Extract the "Last updated" time from the text (after "Last updated: ...")
		if strings.Contains(s.Text(), "Last updated:") {
			lastUpdatedText := strings.TrimSpace(s.Text())
			// Use a regular expression to extract the timestamp from the JavaScript code
			re := regexp.MustCompile(`new Date\((\d+)\*1000\)`)
			matches := re.FindStringSubmatch(lastUpdatedText)
			if len(matches) > 1 {
				// Convert the extracted timestamp to a time object
				timestamp := matches[1]
				if ts, err := convertTimestampToTime(timestamp); err == nil {
					codes.LastUpdated = ts
				} else {
					log.Println("Error converting timestamp:", err)
				}
			}
		}
	})

	// Create a new JSON encoder and print the result to standard output
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ") // For pretty printing the JSON output
	if err := enc.Encode(codes); err != nil {
		log.Fatal(err)
	}
}

// convertTimestampToTime converts a Unix timestamp (in seconds) to a readable date.
func convertTimestampToTime(timestamp string) (string, error) {
	// Convert the timestamp string to an integer
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", err
	}

	// Convert the Unix timestamp to a time.Time object
	t := time.Unix(ts, 0)
	// Format the time as a string in a human-readable format
	return t.Format("01/02/2006, 03:04:05 PM"), nil
}

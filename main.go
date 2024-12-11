package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/proxy"
)

const (
	targetURL     = "https://www.falloutbuilds.com/fo76/nuke-codes/"
	userAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	timeFormat    = "01/02/2006, 03:04:05 PM"
	maxRetries    = 15               // Maximum number of retries
	retryDelay    = 10 * time.Second // Delay between retries
	workflowOwner = "cybellereaper"
	workflowRepo  = "fo76-nuke-codes-fetcher"
	workflowID    = "nuke-codes.yml" // The file name of the workflow you want to trigger
)

// NuclearCodes represents the structure of the nuke codes and related data
type NuclearCodes struct {
	Alpha       string `json:"alpha"`
	Bravo       string `json:"bravo"`
	Charlie     string `json:"charlie"`
	ValidFrom   string `json:"valid_from"`
	ValidTo     string `json:"valid_to"`
	LastUpdated string `json:"last_updated"`
}

// HttpC represents a custom HTTP client with a SOCKS5 proxy setup
type HttpC struct {
	Client *http.Client
}

// NewClient creates a new HTTP client configured to route traffic through Tor's SOCKS5 proxy
func NewClient() *HttpC {
	// Setup Tor SOCKS5 proxy
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		log.Fatalf("Failed to create SOCKS5 proxy: %v", err)
	}

	// Create an HTTP client with the SOCKS5 proxy
	return &HttpC{
		Client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: &http.Transport{Dial: dialer.Dial},
		},
	}
}

// fetchDocument attempts to fetch the webpage and returns a parsed GoQuery document with retry logic
func fetchDocument() (*goquery.Document, error) {
	client := NewClient()

	var doc *goquery.Document
	var err error

	// Retry logic
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("User-Agent", userAgent)
		resp, err := client.Client.Do(req)
		if err != nil {
			log.Printf("Error on attempt %d: %v. Retrying...\n", i+1, err)
			time.Sleep(retryDelay)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Unexpected status code: %d. Retrying...\n", resp.StatusCode)
			time.Sleep(retryDelay)
			continue
		}

		doc, err = goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Printf("Error parsing document on attempt %d: %v. Retrying...\n", i+1, err)
			time.Sleep(retryDelay)
			continue
		}

		// Successfully fetched the document
		return doc, nil
	}

	// After all retries, return the last error encountered
	return nil, fmt.Errorf("failed to fetch document after %d retries: %v", maxRetries, err)
}

// extractNukeCodes extracts the nuclear codes from the document
func extractNukeCodes(doc *goquery.Document) *NuclearCodes {
	codes := &NuclearCodes{}
	doc.Find(".d-flex.flex-column.flex-lg-row.justify-content-lg-around.text-center.h3.mb-0 .text-nowrap").Each(func(_ int, s *goquery.Selection) {
		codeText := strings.TrimSpace(s.Text())
		switch {
		case strings.Contains(codeText, "ALPHA"):
			codes.Alpha = strings.ReplaceAll(codeText, "ALPHA", "")
		case strings.Contains(codeText, "BRAVO"):
			codes.Bravo = strings.ReplaceAll(codeText, "BRAVO", "")
		case strings.Contains(codeText, "CHARLIE"):
			codes.Charlie = strings.ReplaceAll(codeText, "CHARLIE", "")
		}
	})
	return codes
}

// extractValidityTimes extracts the validity period and last updated information from the document
func extractValidityTimes(doc *goquery.Document, codes *NuclearCodes) {
	doc.Find(".small.mb-4").Each(func(_ int, s *goquery.Selection) {
		s.Find("strong").Each(func(i int, strong *goquery.Selection) {
			text := strings.TrimSpace(strong.Text())
			switch i {
			case 0:
				codes.ValidFrom = text
			case 1:
				codes.ValidTo = text
			}
		})

		if strings.Contains(s.Text(), "Last updated:") {
			if timestamp := extractTimestamp(s.Text()); timestamp != "" {
				if ts, err := convertTimestampToTime(timestamp); err == nil {
					codes.LastUpdated = ts
				}
			}
		}
	})
}

// extractTimestamp extracts the timestamp from the JavaScript date
func extractTimestamp(text string) string {
	re := regexp.MustCompile(`new Date\((\d+)\*1000\)`)
	if matches := re.FindStringSubmatch(text); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// convertTimestampToTime converts a timestamp to a formatted time string
func convertTimestampToTime(timestamp string) (string, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", err
	}
	return time.Unix(ts, 0).Format(timeFormat), nil
}

// checkForEmptyFields checks if any of the crucial fields are empty
func checkForEmptyFields(codes *NuclearCodes) bool {
	return codes.Alpha == "" || codes.Bravo == "" || codes.Charlie == "" || codes.ValidFrom == "" || codes.ValidTo == "" || codes.LastUpdated == ""
}

// triggerGitHubAction triggers a new GitHub Actions workflow run
func triggerGitHubAction() error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GitHub token is not set in environment")
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches", workflowOwner, workflowRepo, workflowID)
	data := map[string]interface{}{
		"ref": "main", // Or use another branch or tag
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to trigger GitHub Action: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to trigger GitHub Action. Status: %d", resp.StatusCode)
	}

	log.Println("GitHub Action triggered successfully.")
	return nil
}

func main() {
	var codes *NuclearCodes
	var doc *goquery.Document
	var err error

	// Fetch the document with retry logic
	for i := 0; i < maxRetries; i++ {
		doc, err = fetchDocument()
		if err != nil {
			log.Fatalf("Failed to fetch document: %v", err)
		}

		// Extract nuke codes and validity times
		codes = extractNukeCodes(doc)
		extractValidityTimes(doc, codes)

		// Check if any crucial field is empty
		if checkForEmptyFields(codes) {
			log.Println("Empty field detected, forcing retry and triggering GitHub Action...")

			// Trigger GitHub action to retry the job
			if err := triggerGitHubAction(); err != nil {
				log.Printf("Failed to trigger GitHub Action: %v", err)
			}

			continue
		}

		// If all values are populated, break out of the retry loop
		break
	}

	// Print the extracted data as formatted JSON
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(codes); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}
}

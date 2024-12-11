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
)

const (
	targetURL  = "https://www.falloutbuilds.com/fo76/nuke-codes/"
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	timeFormat = "01/02/2006, 03:04:05 PM"
)

type NuclearCodes struct {
	Alpha       string `json:"alpha"`
	Bravo       string `json:"bravo"`
	Charlie     string `json:"charlie"`
	ValidFrom   string `json:"valid_from"`
	ValidTo     string `json:"valid_to"`
	LastUpdated string `json:"last_updated"`
}

func fetchDocument() (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

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

func extractTimestamp(text string) string {
	re := regexp.MustCompile(`new Date\((\d+)\*1000\)`)
	if matches := re.FindStringSubmatch(text); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func convertTimestampToTime(timestamp string) (string, error) {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", err
	}
	return time.Unix(ts, 0).Format(timeFormat), nil
}

func main() {
	doc, err := fetchDocument()
	if err != nil {
		log.Fatalf("Failed to fetch document: %v", err)
	}

	codes := extractNukeCodes(doc)
	extractValidityTimes(doc, codes)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(codes); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}
}

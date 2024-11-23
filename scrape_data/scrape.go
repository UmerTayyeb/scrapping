// Package scrape handles scraping-related tasks.
// It first scrapes articles from the main page of the main URL,
// then scrapes data from each article and calls Insert_article_metadata
// to store the data in the database.
package scrape

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	mongodb "scrapping/Mongodb"
	"scrapping/models"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/joho/godotenv"
)

// Scrape_articles scrapes articles from a specified URL
// and inserts metadata into the database.
// URL and allowed domains are loaded from environment variables (TARGET_URL, ALLOWED_DOMAINS).
func Scrape_articles() {
	var metadata models.Article_meta_data

	var wg sync.WaitGroup // Create a WaitGroup to synchronize goroutines

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v\n", err)
	}
	mainURL := os.Getenv("TARGET_URL")
	if mainURL == "" {
		log.Fatalf("MAIN_URL not set in .env file\n")
	}

	fmt.Printf("Main URL: %s\n", mainURL)

	// Load allowed domains from environment variable
	allowedDomains := os.Getenv("ALLOWED_DOMAINS")
	if allowedDomains == "" {
		log.Fatalf("ALLOWED_DOMAINS not set in .env file\n")
	}

	// Split allowed domains into a slice
	allowedDomainsSlice := strings.Split(allowedDomains, ",")
	fmt.Printf("Allowed domains: %v\n", allowedDomainsSlice)

	// Create collectors for the main page
	mainPageCollector := colly.NewCollector(colly.AllowedDomains(allowedDomainsSlice...))

	// Create a collector for individual articles
	articleCollector := colly.NewCollector(colly.AllowedDomains(allowedDomainsSlice...))

	// Set User-Agent for both collectors
	mainPageCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	})
	articleCollector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	})

	// Extract article links with hyper refrence from the main page
	mainPageCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if isArticleLink(link) { //check if the link is of news article.
			// Absolute URL for the article
			articleURL := e.Request.AbsoluteURL(link)
			fmt.Println("\n\n\nFound article:", articleURL)
			// Visit the article page
			articleCollector.Visit(articleURL)
		}
	})

	var title string
	// Set the title when scraping individual articles
	articleCollector.OnHTML("title", func(e *colly.HTMLElement) {
		if title == "" {
			title = e.Text
			metadata.Title = title
			fmt.Println("\nTitle:", title)
		}
	})

	// Parse JSON content to get Pub-date, articleURL
	articleCollector.OnHTML("script[type='application/ld+json']", func(e *colly.HTMLElement) {
		// Unmarshal JSON data and extract article info
		scriptContent := e.Text
		var jsonData map[string]interface{}

		if err := json.Unmarshal([]byte(scriptContent), &jsonData); err == nil {
			if datePublished, ok := jsonData["datePublished"].(string); ok {
				// Parse the date time.time into string
				parsedDate, err := time.Parse(time.RFC3339, datePublished)
				if err == nil {
					metadata.DatePublished = parsedDate.Format("2006-01-02")
					fmt.Println("Publication Date (from JSON):", metadata.DatePublished)
				} else {
					fmt.Println("Error parsing date:", err)
				}
			} else {
				fmt.Println("Cant retrieve pub date")
			}

			// Retrieve the Article URL
			if articleURL, ok := jsonData["mainEntityOfPage"].(string); ok {
				fmt.Println("Article URL:", articleURL)
				metadata.ArticleURL = articleURL
			}
		} else {
			fmt.Println("Error unmarshalling JSON:", err)
		}
	})

	// Reset title for each new article scrape
	articleCollector.OnScraped(func(r *colly.Response) {
		title = ""
	})

	// Scrape article content & scrapped time and insert into MongoDB
	articleCollector.OnHTML("article", func(e *colly.HTMLElement) {
		metadata.Content = e.Text
		metadata.ScrapedAt = time.Now() // Set the scrapedAt time here
		fmt.Println("\nContent:\n", e.Text)
		// go mongodb.Insert_article_metadata(metadata)
		wg.Add(1) // Increment the counter for the new goroutine
		go func(meta models.Article_meta_data) {
			defer wg.Done() // Decrement the counter when the goroutine completes
			mongodb.Insert_article_metadata(meta)
		}(metadata)
	})

	// Start scraping from the main page
	fmt.Printf("Visiting main page: %s\n", mainURL)
	if err := mainPageCollector.Visit(mainURL); err != nil {
		log.Fatalf("Failed to visit main page %s: %v\n", mainURL, err)
	}
	wg.Wait() // Wait for goroutine to finsh
}

// isArticleLink checks if the provided link is an article link.
// Input: A string link representing the URL.
// Output: true if the link matches the article pattern, false otherwise.
func isArticleLink(link string) bool {
	if len(link) > 0 && containsSubstring(link, "/news/articles/") {
		return true
	}
	return false
}

// containsSubstring checks if a substring exists in a string.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && stringContains(s, substr)
}

// stringContains checks for an exact substring match.
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

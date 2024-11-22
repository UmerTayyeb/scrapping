// The main package orchestrates the user interaction and main functionality of the application,
// including scraping articles from the web or retrieving them from the MongoDB database.
// It calls the connect func to connect to the MongoDB database
// and prompts the user to either scrape new articles or retrieve existing articles.
// It uses the scrape and mongodb packages to perform these tasks.
package main

import (
	"fmt"
	mongodb "scrapping/Mongodb"
	scrape "scrapping/scrape_data"

	"go.mongodb.org/mongo-driver/bson"
)

// It handles user interaction and executes the two main functionalities: scraping articles or
// retrieving them from the database.
// User input (1 or 2) to decide between scraping or retrieving articles.
func main() {
	// Start of the program
	fmt.Println("starting...")

	// Connecting to the MongoDB database
	err := mongodb.Connect_db()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer mongodb.Disconnect_db() // Disconnect DB after operations

	var input int

	// Prompt user for action
	fmt.Println(`What you want to do?
	Enter Number
	1: Scrape articles
	2: Get articles from DB`)

	// Read the user input
	fmt.Scanln(&input)

	switch input {
	case 1:
		// Scrape articles if user selects option 1
		scrape.Scrape_articles()
	case 2:
		// Allow user to input a filter for retrieving articles from the database
		var date string
		fmt.Println("Enter date (format: YYYY-MM-DD) to fetch articles, or press Enter to fetch all:")
		fmt.Scanln(&date)
		fmt.Println("date parsed:", date)

		var filter bson.M //filter for fetching specific articles from DB
		if date != "" {
			filter = bson.M{"datePublished": date} // Filter by date if provided
		} else {
			filter = bson.M{} // No filter, fetch all articles
		}
		fmt.Println("filter created", filter)

		// Retrieve articles from the database based on the filter
		articles, err := mongodb.Retrieve_articles(filter) //Saving to articles which is a slice
		if err != nil {
			fmt.Printf("Error retrieving articles: %v", err)
		}

		// Print the fetched articles
		for _, article := range articles {
			fmt.Println("Title:", article.Title)
			fmt.Println("Published Date:", article.DatePublished)
			fmt.Println("Scraped at:", article.ScrapedAt)
			fmt.Println("URL:", article.ArticleURL)
			fmt.Println("Content:", article.Content)
			fmt.Println("\n---------\n")
		}
	}
}

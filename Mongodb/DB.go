// Package mongodb manages database-related tasks such as connecting to the database,
// disconnecting, inserting articles and retrieving articles.
package mongodb

import (
	"context"
	"fmt"
	"log"
	"scrapping/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"go.mongodb.org/mongo-driver/mongo/options"
)

var articleCollection *mongo.Collection

// Connect_db connects to MongoDB using the provided URL
// and sets the articleCollection variable for future operations.
// Returns nil on success or an error if the connection fails.
func Connect_db() error {
	fmt.Println("Connecting DB")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return err
	}
	// defer client.Disconnect(context.Background())

	// Set up the article collection for querying and inserting articles
	articleCollection = client.Database("newsdb").Collection("articles")
	fmt.Println("DB connected")
	return nil
}

// Disconnect_db disconnects from MongoDB.
func Disconnect_db() {
	if articleCollection != nil {
		client := articleCollection.Database().Client()
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v\n", err)
		}
	}
}

// Insert_article_metadata inserts article metadata into the MongoDB collection.
// Input: A struct containing the article’s metadata (Title, DatePublished, ArticleURL, Content, ScrapedAt).
func Insert_article_metadata(metadata models.Article_meta_data) {
	var existingArticle bson.M

	// Check if the article with the same title already exists in the database
	article := articleCollection.FindOne(context.Background(), bson.M{"title": metadata.Title}).Decode(&existingArticle)
	if article == nil {
		// Article with the same title already exists, so we skip the insertion
		fmt.Println("\n\nArticle with the same title already exists in the database. Skipping insertion.")
		return
	} else if article != mongo.ErrNoDocuments {
		log.Fatalf("Failed to check for existing article: %v\n", article)
	}

	// If no existing article is found, proceed with insertion
	doc := bson.M{
		"title":         metadata.Title,
		"datePublished": metadata.DatePublished,
		"articleURL":    metadata.ArticleURL,
		"Content":       metadata.Content,
		"scrapedAt":     metadata.ScrapedAt,
	}

	// Insert the document into the collection
	_, err := articleCollection.InsertOne(context.Background(), doc)
	if err != nil {
		log.Fatalf("Failed to insert article into MongoDB: %v\n", err)
	} else {
		fmt.Println("Article metadata inserted into MongoDB successfully!")
	}
}

// Retrieve_articles retrieves articles based on a filter from the database.
// Input: A MongoDB filter (bson.M) for querying specific articles
func Retrieve_articles(filter bson.M) ([]models.Article_meta_data, error) {
	// Set up an empty slice to hold the articles
	var articles []models.Article_meta_data

	// Find the articles matching the filter
	cursor, err := articleCollection.Find(context.Background(), filter)
	if err != nil {
		fmt.Println("cant find article for this date.")
		return nil, fmt.Errorf("failed to retrieve articles: %v", err)
	}
	defer cursor.Close(context.Background())

	// Iterate through the cursor and append each article to the slice
	for cursor.Next(context.Background()) {
		var article models.Article_meta_data
		if err := cursor.Decode(&article); err != nil {
			return nil, fmt.Errorf("failed to decode article: %v", err)
		}
		articles = append(articles, article)
	}

	// Check if any error occurred during iteration
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor iteration error: %v", err)
	}

	// Return the list of articles
	return articles, nil
}

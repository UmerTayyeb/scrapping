// The models package would define the Article_meta_data struct used to represent article data.
package models

import "time"

// The models package would define the Article_meta_data struct used to represent article data.
type Article_meta_data struct {
	DatePublished string    `json:"datePublished"`
	ArticleURL    string    `json:"mainEntityOfPage"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ScrapedAt     time.Time `json:"scrapedAt"`
}

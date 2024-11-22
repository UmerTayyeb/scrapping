package models

import "time"

type Article_meta_data struct {
	DatePublished string    `json:"datePublished"`
	ArticleURL    string    `json:"mainEntityOfPage"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ScrapedAt     time.Time `json:"scrapedAt"`
}

package main

// go:generate easytags $GOFILE json,xml,sql

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mmcdole/gofeed"
)

type RSSFeed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	Link        string `json:"link"`
	Content     string `json:"content"`
}

func main() {

	// create router
	r := gin.Default()
	r.GET("/feeds", feeds)
	r.Run()
}

func feeds(c *gin.Context) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL("https://www.secureworks.com/rss?feed=resources")
	if err != nil {
		log.Printf("Fetching rss feed failed: %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
	}
	items := feed.Items

	feeds := []*RSSFeed{}
	for _, v := range items {
		feed := &RSSFeed{
			Title:       v.Title,
			Description: v.Description,
			Link:        v.Link,
			Content:     v.Content,
		}
		feeds = append(feeds, feed)
	}

	c.JSON(http.StatusOK, feeds)
}

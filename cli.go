package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

func runCLI() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	feeds := getDefaultFeeds()
	summarizer := NewGroqSummarizer(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("📰 RSS Digest - Fetching feeds...")
	fmt.Printf("Feeds: %d sources\n\n", len(feeds))

	var allItems []FeedItem
	for _, feedURL := range feeds {
		items, err := FetchFeed(ctx, feedURL)
		if err != nil {
			log.Printf("Error fetching %s: %v", feedURL, err)
			continue
		}
		allItems = append(allItems, items...)
		fmt.Printf("✓ Fetched %d items from %s\n", len(items), feedURL)
	}

	if len(allItems) == 0 {
		log.Fatal("No items fetched")
	}

	fmt.Printf("\n📊 Total items: %d\n", len(allItems))
	fmt.Println("🤖 Summarizing with Groq...")

	summary, err := summarizer.Summarize(ctx, allItems)
	if err != nil {
		log.Fatalf("Error summarizing: %v", err)
	}

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📰 DAILY DIGEST")
	fmt.Println(strings.Repeat("═", 60) + "\n")
	fmt.Println(summary)
	fmt.Println("\n" + strings.Repeat("═", 60))
}

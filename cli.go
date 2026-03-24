package main

import (
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

	digest, err := generateDigest(apiKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if digest == nil {
		log.Fatal("No items fetched")
	}

	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Println("📰 DAILY DIGEST")
	fmt.Println(strings.Repeat("═", 60) + "\n")
	fmt.Println(digest.Content)
	fmt.Println("\n" + strings.Repeat("═", 60))
}

func runGenerate() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}

	digest, err := generateDigest(apiKey)
	if err != nil {
		log.Fatalf("Error generating digest: %v", err)
	}

	if digest == nil {
		log.Fatal("No items fetched")
	}

	storage := NewStorage(dataDir)
	if err := storage.Save(digest.Content, digest.ItemCount); err != nil {
		log.Fatalf("Error saving digest: %v", err)
	}

	log.Printf("✅ Digest saved (%d items, %s)", digest.ItemCount, digest.Timestamp.Format(time.RFC3339))
}

func runHelp() {
	fmt.Println("RSS Digest - Daily news summarizer")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  rss-digest              Run in CLI mode (print to stdout)")
	fmt.Println("  rss-digest --serve      Start web server")
	fmt.Println("  rss-digest --generate   Generate digest and save to storage")
	fmt.Println("  rss-digest --help       Show this help")
	fmt.Println("")
	fmt.Println("Environment:")
	fmt.Println("  GROQ_API_KEY   Groq API key (required)")
	fmt.Println("  DATA_DIR       Storage directory (default: /data)")
	fmt.Println("  PORT           Server port (default: 8080)")
}

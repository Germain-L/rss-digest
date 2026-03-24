package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

var storage *Storage

func startServer() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}
	storage = NewStorage(dataDir)

	fs := http.FileServer(http.Dir("frontend"))
	http.Handle("/", fs)
	http.HandleFunc("/api/digest", handleGetDigest)
	http.HandleFunc("/api/refresh", handleRefreshDigest)
	http.HandleFunc("/health", handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Server running at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleGetDigest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if storage == nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	d := storage.Get()
	if d == nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	json.NewEncoder(w).Encode(d)
}

func handleRefreshDigest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		http.Error(w, "GROQ_API_KEY not set", http.StatusInternalServerError)
		return
	}

	digest, err := generateDigest(apiKey)
	if err != nil {
		log.Printf("Error generating digest: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if storage != nil {
		storage.Save(digest.Content, digest.ItemCount)
	}

	json.NewEncoder(w).Encode(digest)
}

func generateDigest(apiKey string) (*StoredDigest, error) {
	feeds := getDefaultFeeds()
	summarizer := NewGroqSummarizer(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	log.Println("📰 Fetching feeds...")
	var allItems []FeedItem
	for _, feedURL := range feeds {
		items, err := FetchFeed(ctx, feedURL)
		if err != nil {
			log.Printf("Error fetching %s: %v", feedURL, err)
			continue
		}
		allItems = append(allItems, items...)
		log.Printf("✓ %s: %d items", feedURL, len(items))
	}

	if len(allItems) == 0 {
		return nil, nil
	}

	log.Printf("📊 Total: %d items", len(allItems))
	log.Println("🤖 Summarizing...")

	summary, err := summarizer.Summarize(ctx, allItems)
	if err != nil {
		return nil, err
	}

	log.Println("✅ Digest generated")

	return &StoredDigest{
		Content:   summary,
		Timestamp: time.Now(),
		ItemCount: len(allItems),
	}, nil
}

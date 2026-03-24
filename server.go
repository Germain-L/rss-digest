package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type CachedDigest struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	digestCache CachedDigest
	cacheMutex  sync.RWMutex
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--serve" {
		startServer()
	} else {
		runCLI()
	}
}

func startServer() {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("GROQ_API_KEY environment variable is required")
	}

	// Serve static files
	fs := http.FileServer(http.Dir("frontend"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/digest", handleGetDigest)
	http.HandleFunc("/api/refresh", handleRefreshDigest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🌐 Server running at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleGetDigest(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(digestCache)
}

func handleRefreshDigest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	apiKey := os.Getenv("GROQ_API_KEY")
	feeds := getDefaultFeeds()
	summarizer := NewGroqSummarizer(apiKey)

	ctx, cancel := contextWithTimeout(2 * time.Minute)
	defer cancel()

	var allItems []FeedItem
	for _, feedURL := range feeds {
		items, err := FetchFeed(ctx, feedURL)
		if err != nil {
			log.Printf("Error fetching %s: %v", feedURL, err)
			continue
		}
		allItems = append(allItems, items...)
	}

	if len(allItems) == 0 {
		http.Error(w, "No items fetched", http.StatusInternalServerError)
		return
	}

	summary, err := summarizer.Summarize(ctx, allItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cacheMutex.Lock()
	digestCache = CachedDigest{
		Content:   summary,
		Timestamp: time.Now(),
	}
	cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(digestCache)
}

func contextWithTimeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

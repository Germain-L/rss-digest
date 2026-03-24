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
var configManager *ConfigManager

func startServer() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "/data"
	}
	storage = NewStorage(dataDir)
	configManager = NewConfigManager(dataDir)

	fs := http.FileServer(http.Dir("frontend"))
	http.Handle("/", fs)
	
	// API endpoints
	http.HandleFunc("/api/digest", handleGetDigest)
	http.HandleFunc("/api/refresh", handleRefreshDigest)
	http.HandleFunc("/api/feeds", handleFeeds)
	http.HandleFunc("/api/feeds/add", handleAddFeed)
	http.HandleFunc("/api/feeds/toggle", handleToggleFeed)
	http.HandleFunc("/api/feeds/remove", handleRemoveFeed)
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
		storage.Save(digest.Content, digest.ItemCount, digest.Articles)
	}

	json.NewEncoder(w).Encode(digest)
}

func handleFeeds(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if configManager == nil {
		json.NewEncoder(w).Encode([]FeedConfig{})
		return
	}

	json.NewEncoder(w).Encode(configManager.GetAllFeeds())
}

func handleAddFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var feed FeedConfig
	if err := json.NewDecoder(r.Body).Decode(&feed); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if feed.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if feed.Name == "" {
		feed.Name = feed.URL
	}

	if configManager == nil {
		http.Error(w, "Config not initialized", http.StatusInternalServerError)
		return
	}

	if err := configManager.AddFeed(feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleToggleFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req struct {
		URL     string `json:"url"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if configManager == nil {
		http.Error(w, "Config not initialized", http.StatusInternalServerError)
		return
	}

	if err := configManager.ToggleFeed(req.URL, req.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleRemoveFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if configManager == nil {
		http.Error(w, "Config not initialized", http.StatusInternalServerError)
		return
	}

	if err := configManager.RemoveFeed(req.URL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func generateDigest(apiKey string) (*StoredDigest, error) {
	summarizer := NewGroqSummarizer(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var feeds []string
	if configManager != nil {
		for _, f := range configManager.GetFeeds() {
			feeds = append(feeds, f.URL)
		}
	}
	if len(feeds) == 0 {
		feeds = getDefaultFeeds()
	}

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

	// Convert FeedItems to Articles for storage (limit to 30 most recent)
	articles := make([]Article, 0, 30)
	for i, item := range allItems {
		if i >= 30 {
			break
		}
		articles = append(articles, Article{
			Title:  item.Title,
			Link:   item.Link,
			Source: item.Source,
		})
	}

	return &StoredDigest{
		Content:   summary,
		Timestamp: time.Now(),
		ItemCount: len(allItems),
		Articles:  articles,
	}, nil
}

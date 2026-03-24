package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type FeedConfig struct {
	URL     string `json:"url"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type Config struct {
	Feeds []FeedConfig `json:"feeds"`
}

type ConfigManager struct {
	path     string
	mutex    sync.RWMutex
	config   *Config
}

func NewConfigManager(dataDir string) *ConfigManager {
	path := filepath.Join(dataDir, "config.json")
	cm := &ConfigManager{path: path}
	cm.load()
	return cm
}

func (cm *ConfigManager) load() {
	data, err := os.ReadFile(cm.path)
	if err != nil {
		// Create default config
		cm.config = &Config{Feeds: getDefaultFeedConfigs()}
		cm.save()
		return
	}
	var c Config
	if json.Unmarshal(data, &c) == nil {
		cm.config = &c
	}
}

func (cm *ConfigManager) save() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cm.path, data, 0644)
}

func (cm *ConfigManager) GetFeeds() []FeedConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	var enabled []FeedConfig
	for _, f := range cm.config.Feeds {
		if f.Enabled {
			enabled = append(enabled, f)
		}
	}
	return enabled
}

func (cm *ConfigManager) GetAllFeeds() []FeedConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return cm.config.Feeds
}

func (cm *ConfigManager) AddFeed(feed FeedConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	feed.Enabled = true
	cm.config.Feeds = append(cm.config.Feeds, feed)
	return cm.save()
}

func (cm *ConfigManager) RemoveFeed(url string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	for i, f := range cm.config.Feeds {
		if f.URL == url {
			cm.config.Feeds = append(cm.config.Feeds[:i], cm.config.Feeds[i+1:]...)
			return cm.save()
		}
	}
	return nil
}

func (cm *ConfigManager) ToggleFeed(url string, enabled bool) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	for i := range cm.config.Feeds {
		if cm.config.Feeds[i].URL == url {
			cm.config.Feeds[i].Enabled = enabled
			return cm.save()
		}
	}
	return nil
}

func getDefaultFeedConfigs() []FeedConfig {
	return []FeedConfig{
		{URL: "https://hnrss.org/frontpage", Name: "Hacker News", Enabled: true},
		{URL: "https://feeds.arstechnica.com/arstechnica/index", Name: "Ars Technica", Enabled: true},
		{URL: "https://www.theverge.com/rss/index.xml", Name: "The Verge", Enabled: true},
		{URL: "https://techcrunch.com/feed/", Name: "TechCrunch", Enabled: true},
		{URL: "https://blog.golang.org/feed.atom", Name: "Go Blog", Enabled: true},
		{URL: "https://www.reddit.com/r/golang/.rss", Name: "r/golang", Enabled: true},
		{URL: "https://www.reddit.com/r/programming/.rss", Name: "r/programming", Enabled: true},
		{URL: "https://www.reddit.com/r/MachineLearning/.rss", Name: "r/MachineLearning", Enabled: true},
		{URL: "https://www.reddit.com/r/artificial/.rss", Name: "r/artificial", Enabled: true},
		{URL: "https://www.reddit.com/r/kubernetes/.rss", Name: "r/kubernetes", Enabled: true},
		{URL: "https://www.reddit.com/r/devops/.rss", Name: "r/devops", Enabled: true},
	}
}

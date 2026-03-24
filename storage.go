package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Article struct {
	Title  string `json:"title"`
	Link   string `json:"link"`
	Source string `json:"source"`
}

type StoredDigest struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	ItemCount int       `json:"itemCount"`
	Articles  []Article `json:"articles"`
}

type Storage struct {
	path     string
	mutex    sync.RWMutex
	cache    *StoredDigest
}

func NewStorage(dataDir string) *Storage {
	path := filepath.Join(dataDir, "digest.json")
	s := &Storage{path: path}
	s.load()
	return s
}

func (s *Storage) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var d StoredDigest
	if json.Unmarshal(data, &d) == nil {
		s.cache = &d
	}
}

func (s *Storage) Get() *StoredDigest {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.cache
}

func (s *Storage) Save(content string, itemCount int, articles []Article) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	d := &StoredDigest{
		Content:   content,
		Timestamp: time.Now(),
		ItemCount: itemCount,
		Articles:  articles,
	}

	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return err
	}

	s.cache = d
	return nil
}

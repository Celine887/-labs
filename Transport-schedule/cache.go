package cache

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"yandex-route-finder/models"
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte) error
	GetRoute(request models.RouteRequest) ([]models.CompleteRoute, bool)
	SetRoute(request models.RouteRequest, routes []models.CompleteRoute) error
}

type MemoryCache struct {
	data     map[string][]byte
	routes   map[string][]models.CompleteRoute
	mutex    sync.RWMutex
	fileDir  string
	ttl      time.Duration
	lastSave time.Time
}

func NewMemoryCache(fileDir string, ttl time.Duration) (*MemoryCache, error) {
	cache := &MemoryCache{
		data:     make(map[string][]byte),
		routes:   make(map[string][]models.CompleteRoute),
		fileDir:  fileDir,
		ttl:      ttl,
		lastSave: time.Now(),
	}

	if fileDir != "" {
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}

		if err := cache.loadFromFile(); err != nil {
			return nil, fmt.Errorf("failed to load cache from file: %w", err)
		}
	}

	return cache, nil
}

func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	data, exists := c.data[key]
	return data, exists
}

func (c *MemoryCache) Set(key string, data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = data

	if c.fileDir != "" && time.Since(c.lastSave) > 5*time.Minute {
		if err := c.saveToFile(); err != nil {
			return err
		}
		c.lastSave = time.Now()
	}

	return nil
}

func generateRouteKey(req models.RouteRequest) string {
	return fmt.Sprintf("%s-%s-%s-%d",
		req.FromCity,
		req.ToCity,
		req.Date.Format("2006-01-02"),
		req.MaxTransfers)
}

func (c *MemoryCache) GetRoute(request models.RouteRequest) ([]models.CompleteRoute, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	key := generateRouteKey(request)
	routes, exists := c.routes[key]
	return routes, exists
}

func (c *MemoryCache) SetRoute(request models.RouteRequest, routes []models.CompleteRoute) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := generateRouteKey(request)
	c.routes[key] = routes

	if c.fileDir != "" && time.Since(c.lastSave) > 5*time.Minute {
		if err := c.saveToFile(); err != nil {
			return err
		}
		c.lastSave = time.Now()
	}

	return nil
}

func (c *MemoryCache) saveToFile() error {

	dataCachePath := filepath.Join(c.fileDir, "data_cache.json")
	dataBytes, err := json.Marshal(c.data)
	if err != nil {
		return fmt.Errorf("failed to marshal data cache: %w", err)
	}

	if err := ioutil.WriteFile(dataCachePath, dataBytes, 0644); err != nil {
		return fmt.Errorf("failed to write data cache to file: %w", err)
	}

	routesCachePath := filepath.Join(c.fileDir, "routes_cache.json")
	routesBytes, err := json.Marshal(c.routes)
	if err != nil {
		return fmt.Errorf("failed to marshal routes cache: %w", err)
	}

	if err := ioutil.WriteFile(routesCachePath, routesBytes, 0644); err != nil {
		return fmt.Errorf("failed to write routes cache to file: %w", err)
	}

	return nil
}

func (c *MemoryCache) loadFromFile() error {

	dataCachePath := filepath.Join(c.fileDir, "data_cache.json")
	if _, err := os.Stat(dataCachePath); err == nil {
		dataBytes, err := ioutil.ReadFile(dataCachePath)
		if err != nil {
			return fmt.Errorf("failed to read data cache file: %w", err)
		}

		if err := json.Unmarshal(dataBytes, &c.data); err != nil {
			return fmt.Errorf("failed to unmarshal data cache: %w", err)
		}
	}

	routesCachePath := filepath.Join(c.fileDir, "routes_cache.json")
	if _, err := os.Stat(routesCachePath); err == nil {
		routesBytes, err := ioutil.ReadFile(routesCachePath)
		if err != nil {
			return fmt.Errorf("failed to read routes cache file: %w", err)
		}

		if err := json.Unmarshal(routesBytes, &c.routes); err != nil {
			return fmt.Errorf("failed to unmarshal routes cache: %w", err)
		}
	}

	return nil
}

func (c *MemoryCache) Cleanup() {

}

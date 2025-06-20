package monitor

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/repo-scm/proxy/config"
)

type SiteStatus struct {
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Healthy      bool      `json:"healthy"`
	ResponseTime int64     `json:"responseTime"`
	Connections  int       `json:"connections"`
	QueueSize    int       `json:"queueSize"`
	LastCheck    time.Time `json:"lastCheck"`
	Error        string    `json:"error,omitempty"`
}

type Monitor struct {
	config *config.Config
	sites  map[string]*SiteStatus
	mutex  sync.RWMutex
	client *http.Client
}

func NewMonitor(cfg *config.Config) *Monitor {
	m := &Monitor{
		config: cfg,
		sites:  make(map[string]*SiteStatus),
		client: &http.Client{Timeout: 10 * time.Second},
	}

	// Initialize with detailed mock data
	m.initializeMockData()

	return m
}

func (m *Monitor) initializeMockData() {
	now := time.Now()

	mockSites := []*SiteStatus{
		{
			Name:         "gerrit-production",
			URL:          "https://gerrit.example.com",
			Healthy:      true,
			ResponseTime: 245,
			Connections:  23,
			QueueSize:    5,
			LastCheck:    now.Add(-2 * time.Minute),
			Error:        "",
		},
		{
			Name:         "gerrit-staging",
			URL:          "https://gerrit-staging.example.com",
			Healthy:      true,
			ResponseTime: 180,
			Connections:  8,
			QueueSize:    2,
			LastCheck:    now.Add(-1 * time.Minute),
			Error:        "",
		},
		{
			Name:         "gerrit-dev",
			URL:          "https://gerrit-dev.example.com",
			Healthy:      false,
			ResponseTime: 0,
			Connections:  0,
			QueueSize:    12,
			LastCheck:    now.Add(-5 * time.Minute),
			Error:        "Connection timeout",
		},
		{
			Name:         "code-review-mirror",
			URL:          "https://mirror.gerrit.example.com",
			Healthy:      true,
			ResponseTime: 320,
			Connections:  15,
			QueueSize:    1,
			LastCheck:    now.Add(-30 * time.Second),
			Error:        "",
		},
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, site := range mockSites {
		m.sites[site.Name] = site
	}
}

func (m *Monitor) GetAllSitesStatus() []*SiteStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sites := make([]*SiteStatus, 0, len(m.sites))
	for _, site := range m.sites {
		sites = append(sites, site)
	}

	return sites
}

func (m *Monitor) GetSiteHealth(siteName string) map[string]interface{} {
	m.mutex.RLock()
	site, exists := m.sites[siteName]
	m.mutex.RUnlock()

	if !exists {
		return map[string]interface{}{
			"error": "Site not found",
		}
	}

	return map[string]interface{}{
		"name":         site.Name,
		"healthy":      site.Healthy,
		"responseTime": site.ResponseTime,
		"lastCheck":    site.LastCheck,
		"error":        site.Error,
	}
}

func (m *Monitor) GetSiteQueues(siteName string) map[string]interface{} {
	// Enhanced mock data with more realistic queue information
	queues := map[string]map[string]interface{}{
		"gerrit-production": {
			"receive": map[string]interface{}{
				"pending": 12,
				"running": 3,
			},
			"send-email": map[string]interface{}{
				"pending": 8,
				"running": 2,
			},
			"ssh": map[string]interface{}{
				"pending": 5,
				"running": 1,
			},
			"index": map[string]interface{}{
				"pending": 15,
				"running": 4,
			},
		},
		"gerrit-staging": {
			"receive": map[string]interface{}{
				"pending": 3,
				"running": 1,
			},
			"send-email": map[string]interface{}{
				"pending": 2,
				"running": 0,
			},
			"ssh": map[string]interface{}{
				"pending": 1,
				"running": 0,
			},
			"index": map[string]interface{}{
				"pending": 4,
				"running": 1,
			},
		},
	}

	if data, exists := queues[siteName]; exists {
		return data
	}

	// Default queue data
	return map[string]interface{}{
		"receive": map[string]interface{}{
			"pending": 0,
			"running": 0,
		},
		"send-email": map[string]interface{}{
			"pending": 0,
			"running": 0,
		},
		"ssh": map[string]interface{}{
			"pending": 0,
			"running": 0,
		},
	}
}

func (m *Monitor) GetSiteConnections(siteName string) map[string]interface{} {
	// Enhanced mock data with more realistic connection information
	connections := map[string]map[string]interface{}{
		"gerrit-production": {
			"http": map[string]interface{}{
				"active": 45,
				"total":  2840,
				"peak":   89,
			},
			"ssh": map[string]interface{}{
				"active": 23,
				"total":  1456,
				"peak":   67,
			},
		},
		"gerrit-staging": {
			"http": map[string]interface{}{
				"active": 12,
				"total":  890,
				"peak":   34,
			},
			"ssh": map[string]interface{}{
				"active": 5,
				"total":  234,
				"peak":   18,
			},
		},
		"gerrit-dev": {
			"http": map[string]interface{}{
				"active": 0,
				"total":  156,
				"peak":   12,
			},
			"ssh": map[string]interface{}{
				"active": 0,
				"total":  45,
				"peak":   8,
			},
		},
	}

	if data, exists := connections[siteName]; exists {
		return data
	}

	// Default connection data
	return map[string]interface{}{
		"http": map[string]interface{}{
			"active": 0,
			"total":  0,
			"peak":   0,
		},
		"ssh": map[string]interface{}{
			"active": 0,
			"total":  0,
			"peak":   0,
		},
	}
}

func (m *Monitor) checkAllSites() {
	// This would read from config to get list of all sites
	// For now, use mock data
	mockSites := []struct {
		name string
		url  string
	}{
		{"gerrit-production", "https://gerrit.example.com"},
		{"gerrit-staging", "https://gerrit-staging.example.com"},
		{"gerrit-dev", "https://gerrit-dev.example.com"},
		{"code-review-mirror", "https://mirror.gerrit.example.com"},
	}

	for _, mockSite := range mockSites {
		m.checkSite(mockSite.name, mockSite.url)
	}
}

func (m *Monitor) checkSite(name, url string) {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url+"/config/server/healthcheck~status", nil)
	if err != nil {
		m.updateSiteStatus(name, url, false, 0, err.Error())
		return
	}

	resp, err := m.client.Do(req)
	if err != nil {
		m.updateSiteStatus(name, url, false, 0, err.Error())
		return
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	responseTime := time.Since(start).Milliseconds()
	healthy := resp.StatusCode == 200

	m.updateSiteStatus(name, url, healthy, responseTime, "")
}

func (m *Monitor) updateSiteStatus(name, url string, healthy bool, responseTime int64, errorMsg string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Preserve existing connections and queue data if available
	connections := 15
	queueSize := 8
	if existing, exists := m.sites[name]; exists {
		connections = existing.Connections
		queueSize = existing.QueueSize
	}

	m.sites[name] = &SiteStatus{
		Name:         name,
		URL:          url,
		Healthy:      healthy,
		ResponseTime: responseTime,
		Connections:  connections,
		QueueSize:    queueSize,
		LastCheck:    time.Now(),
		Error:        errorMsg,
	}
}

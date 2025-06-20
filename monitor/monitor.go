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
	return &Monitor{
		config: cfg,
		sites:  make(map[string]*SiteStatus),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (m *Monitor) GetAllSitesStatus() []*SiteStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// If no sites cached, check them now
	if len(m.sites) == 0 {
		m.checkAllSites()
	}

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
	// This would typically call REST API to get queue information
	// For now, return mock data
	return map[string]interface{}{
		"receive": map[string]interface{}{
			"pending": 5,
			"running": 2,
		},
		"send-email": map[string]interface{}{
			"pending": 0,
			"running": 1,
		},
		"ssh": map[string]interface{}{
			"pending": 3,
			"running": 0,
		},
	}
}

func (m *Monitor) GetSiteConnections(siteName string) map[string]interface{} {
	// This would typically call REST API to get connection information
	// For now, return mock data
	return map[string]interface{}{
		"http": map[string]interface{}{
			"active": 15,
			"total":  1243,
			"peak":   45,
		},
		"ssh": map[string]interface{}{
			"active": 8,
			"total":  892,
			"peak":   23,
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
		{"gerrit-main", "https://gerrit.example.com"},
		{"gerrit-staging", "https://gerrit-staging.example.com"},
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

	m.sites[name] = &SiteStatus{
		Name:         name,
		URL:          url,
		Healthy:      healthy,
		ResponseTime: responseTime,
		Connections:  15, // Mock data
		QueueSize:    8,  // Mock data
		LastCheck:    time.Now(),
		Error:        errorMsg,
	}
}

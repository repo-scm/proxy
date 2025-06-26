package monitor

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/repo-scm/proxy/config"
)

const (
	siteName = "gerrit"

	ConnectionMax = 65536
	QueueMax      = 65536
	Weight        = 10
)

type SiteStatus struct {
	Name         string    `json:"name"`
	Location     string    `json:"location"`
	Url          string    `json:"url"`
	Host         string    `json:"host"`
	Healthy      bool      `json:"healthy"`
	ResponseTime int64     `json:"responseTime"`
	Connections  int       `json:"connections"`
	QueueSize    int       `json:"queueSize"`
	Score        int       `json:"score"`
	LastCheck    time.Time `json:"lastCheck"`
	Error        string    `json:"error"`
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

	m.initializeMonitor()

	return m
}

func (m *Monitor) initializeMonitor() {
	for key, val := range m.config.Gerrits {
		m.sites[key] = &SiteStatus{
			Name:     key,
			Location: val.Location,
			Url:      val.Http.Url,
			Host:     val.Ssh.Host,
		}
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

func (m *Monitor) GetSiteHealth(name string) map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	responseTime, err := m.getResponseTime(name)
	if err != nil {
		return map[string]interface{}{
			"name":         name,
			"healthy":      false,
			"responseTime": responseTime,
			"lastCheck":    time.Now(),
			"error":        fmt.Sprintf("failed to get health for site %s", name),
		}
	}

	return map[string]interface{}{
		"name":         name,
		"healthy":      true,
		"responseTime": responseTime,
		"lastCheck":    time.Now(),
		"error":        "",
	}
}

func (m *Monitor) GetSiteQueues(name string) map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	queue, err := m.getQueue(name)
	if err != nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"queueSize": queue,
	}
}

func (m *Monitor) GetSiteConnections(name string) map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	connections, err := m.getConnection(name)
	if err != nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"connections": connections,
	}
}

func (m *Monitor) GetAvailableSite() (*SiteStatus, error) {
	var bestSite *SiteStatus

	if len(m.sites) == 0 {
		return nil, errors.New("no sites available\n")
	}

	siteChan := make(chan *SiteStatus, len(m.sites))
	activeGoroutines := 0

	for _, site := range m.sites {
		if bestSite != nil {
			if bestSite.Name == site.Name {
				continue
			}
		} else {
			bestSite = site
		}
		activeGoroutines++
		go func(site *SiteStatus) {
			siteChan <- m.getSiteStatus(site.Name)
		}(site)
	}

	bestScore := ConnectionMax*Weight + QueueMax + 1000 // High fallback score

	for i := 0; i < activeGoroutines; i++ {
		site := <-siteChan
		if site.Error == "" && site.Score < bestScore {
			bestSite = site
			bestScore = site.Score
		}
	}

	return bestSite, nil
}

func (m *Monitor) getResponseTime(name string) (float64, error) {
	cmd := exec.Command("ssh", "-p", strconv.Itoa(m.config.Gerrits[name].Ssh.Port), "-i", m.config.Gerrits[name].Ssh.Key, "-o", "ConnectTimeout=5", fmt.Sprintf("%s@%s", m.config.Gerrits[name].Ssh.User, m.config.Gerrits[name].Ssh.Host), siteName, "version")

	start := time.Now()
	err := cmd.Run()
	elapsed := time.Since(start)

	if err != nil {
		return 1000.0, err // High penalty for unreachable sites
	}

	return float64(elapsed.Nanoseconds()) / 1000000.0, nil // Convert to milliseconds
}

func (m *Monitor) getSiteStatus(name string) *SiteStatus {
	type result struct {
		connections int
		queue       int
		connErr     error
		queueErr    error
	}

	ch := make(chan result, 1)

	go func() {
		connections, connErr := m.getConnection(name)
		queue, queueErr := m.getQueue(name)

		ch <- result{
			connections: connections,
			queue:       queue,
			connErr:     connErr,
			queueErr:    queueErr,
		}
	}()

	res := <-ch
	if res.connErr != nil || res.queueErr != nil {
		return &SiteStatus{
			Name:         name,
			Location:     m.sites[name].Location,
			Url:          m.sites[name].Url,
			Host:         m.sites[name].Host,
			Healthy:      false,
			ResponseTime: -1,
			Connections:  ConnectionMax,
			QueueSize:    QueueMax,
			Score:        -1,
			LastCheck:    time.Now(),
			Error:        fmt.Sprintf("failed to get status for site %s", name),
		}
	}

	responseTime, _ := m.getResponseTime(name)
	score := m.calculateScore(name, res.connections, res.queue)

	return &SiteStatus{
		Name:         name,
		Location:     m.sites[name].Location,
		Url:          m.sites[name].Url,
		Host:         m.sites[name].Host,
		Healthy:      true,
		ResponseTime: int64(responseTime),
		Connections:  res.connections,
		QueueSize:    res.queue,
		Score:        score,
		LastCheck:    time.Now(),
		Error:        "",
	}
}

func (m *Monitor) getQueue(name string) (int, error) {
	cmd := exec.Command("ssh", "-p", strconv.Itoa(m.config.Gerrits[name].Ssh.Port), "-i", m.config.Gerrits[name].Ssh.Key, fmt.Sprintf("%s@%s", m.config.Gerrits[name].Ssh.User, m.config.Gerrits[name].Ssh.Host), siteName, "version")
	if err := cmd.Run(); err != nil {
		return QueueMax, nil
	}

	cmd = exec.Command("ssh", "-p", strconv.Itoa(m.config.Gerrits[name].Ssh.Port), "-i", m.config.Gerrits[name].Ssh.Key, fmt.Sprintf("%s@%s", m.config.Gerrits[name].Ssh.User, m.config.Gerrits[name].Ssh.Host), siteName, "show-queue", "-w")
	output, err := cmd.Output()
	if err != nil {
		return QueueMax, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "tasks") && !strings.Contains(line, "waiting") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				if queue, err := strconv.Atoi(fields[0]); err == nil {
					return queue, nil
				}
			}
		}
	}

	return QueueMax, nil
}

func (m *Monitor) getConnection(name string) (int, error) {
	cmd := exec.Command("ssh", "-p", strconv.Itoa(m.config.Gerrits[name].Ssh.Port), "-i", m.config.Gerrits[name].Ssh.Key, fmt.Sprintf("%s@%s", m.config.Gerrits[name].Ssh.User, m.config.Gerrits[name].Ssh.Host), siteName, "version")
	if err := cmd.Run(); err != nil {
		return ConnectionMax, nil
	}

	cmd = exec.Command("ssh", "-p", strconv.Itoa(m.config.Gerrits[name].Ssh.Port), "-i", m.config.Gerrits[name].Ssh.Key, fmt.Sprintf("%s@%s", m.config.Gerrits[name].Ssh.User, m.config.Gerrits[name].Ssh.Host), siteName, "show-connections", "-w")
	output, err := cmd.Output()
	if err != nil {
		return ConnectionMax, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "connections") {
			parts := strings.Split(line, ":")
			if len(parts) > 0 {
				fields := strings.Fields(parts[0])
				if len(fields) > 0 {
					if connection, err := strconv.Atoi(fields[0]); err == nil {
						return connection, nil
					}
				}
			}
		}
	}

	return ConnectionMax, nil
}

func (m *Monitor) getLatencyPenalty(name string) int {
	// Add penalty based on network latency for remote sites
	latency, _ := m.getResponseTime(name)

	// Convert latency to penalty (higher latency = higher penalty)
	// Latency in milliseconds, penalty multiplier
	penalty := int(latency * 0.1) // 10ms latency = 1 point penalty

	return penalty
}

func (m *Monitor) getConnectionEfficiency(connections int) int {
	// Lower connections are better, but add diminishing returns
	if connections == 0 {
		return -5 // Bonus for no connections
	} else if connections <= 5 {
		return -2 // Small bonus for low connections
	} else if connections <= 10 {
		return 0 // Neutral
	} else {
		return connections / 5 // Penalty increases with high connections
	}
}

func (m *Monitor) getQueueEfficiency(queue int) int {
	// Lower queue is better, exponential penalty for high queues
	if queue == 0 {
		return -10 // Significant bonus for empty queue
	} else if queue <= 5 {
		return -3 // Small bonus for low queue
	} else if queue <= 20 {
		return queue / 4 // Moderate penalty
	} else {
		return queue / 2 // Higher penalty for large queues
	}
}

func (m *Monitor) calculateScore(name string, connections, queue int) int {
	// Calculate base score using the original Weight constant
	baseScore := connections*Weight + queue

	latencyPenalty := m.getLatencyPenalty(name)
	connectionEfficiency := m.getConnectionEfficiency(connections)
	queueEfficiency := m.getQueueEfficiency(queue)

	// Get the importance weight from the config for this specific site (0.0 to 1.0)
	siteImportance := m.config.Gerrits[name].Weight
	if siteImportance == 0 {
		siteImportance = 1.0 // Default to full importance if not configured
	}

	// Calculate total score and apply site importance as a multiplier
	// Higher importance (closer to 1.0) = lower final score (more preferred)
	// Lower importance (closer to 0.0) = higher final score (less preferred)
	totalScore := baseScore + latencyPenalty + connectionEfficiency + queueEfficiency

	// Apply importance factor: invert it so higher importance gives lower score
	finalScore := int(float32(totalScore) / siteImportance)

	return finalScore
}

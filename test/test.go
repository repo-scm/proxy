package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

const (
	siteName = "gerrit"

	ConnectionMax = 65536
	QueueMax      = 65536
	Weight        = 10
)

var (
	app = kingpin.New("test", "test")

	sites = app.Flag("sites", "sites list").Short('s').String()
)

type Monitor struct {
	Host string
	Site string
}

type Metrics struct {
	Site        string
	Connections int
	Queue       int
	Score       int
	Error       error
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) getLocal() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}

	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func (m *Monitor) getQueue(host string) (int, error) {
	cmd := exec.Command("ssh", "-p", "29418", host, siteName, "version")
	if err := cmd.Run(); err != nil {
		return QueueMax, nil
	}

	cmd = exec.Command("ssh", "-p", "29418", host, siteName, "show-queue", "-w")
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

func (m *Monitor) getConnection(host string) (int, error) {
	cmd := exec.Command("ssh", "-p", "29418", host, siteName, "version")
	if err := cmd.Run(); err != nil {
		return ConnectionMax, nil
	}

	cmd = exec.Command("ssh", "-p", "29418", host, siteName, "show-connections", "-w")
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

func (m *Monitor) getLatencyPenalty(site string) int {
	helper := func(site string) float64 {
		cmd := exec.Command("ssh", "-p", "29418", "-o", "ConnectTimeout=5", site, siteName, "version")
		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)
		if err != nil {
			return 1000.0 // High penalty for unreachable sites
		}
		return float64(elapsed.Nanoseconds()) / 1000000.0 // Convert to milliseconds
	}

	// Add penalty based on network latency for remote sites
	latency := helper(site)

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

func (m *Monitor) calculateScore(site string, connections, queue int) int {
	baseScore := connections*Weight + queue

	latencyPenalty := m.getLatencyPenalty(site)
	connectionEfficiency := m.getConnectionEfficiency(connections)
	queueEfficiency := m.getQueueEfficiency(queue)

	totalScore := baseScore + latencyPenalty + connectionEfficiency + queueEfficiency

	return totalScore
}

func (m *Monitor) getMetrics(site string) Metrics {
	type result struct {
		connections int
		queue       int
		connErr     error
		queueErr    error
	}

	ch := make(chan result, 1)

	go func() {
		connections, connErr := m.getConnection(site)
		queue, queueErr := m.getQueue(site)

		ch <- result{
			connections: connections,
			queue:       queue,
			connErr:     connErr,
			queueErr:    queueErr,
		}
	}()

	res := <-ch
	if res.connErr != nil || res.queueErr != nil {
		return Metrics{
			Site:        site,
			Connections: ConnectionMax,
			Queue:       QueueMax,
			Score:       m.calculateScore(site, ConnectionMax, QueueMax),
			Error:       fmt.Errorf("failed to get metrics for site %s", site),
		}
	}

	score := m.calculateScore(site, res.connections, res.queue)

	return Metrics{
		Site:        site,
		Connections: res.connections,
		Queue:       res.queue,
		Score:       score,
		Error:       nil,
	}
}

func (m *Monitor) runTest(sites []string) (string, error) {
	if len(sites) == 0 {
		return "", fmt.Errorf("no sites provided")
	}

	bestSite := m.Site
	if bestSite == "" && len(sites) > 0 {
		bestSite = sites[0]
	}

	var bestScore int
	if bestSite != "" {
		bestMetrics := m.getMetrics(bestSite)
		bestScore = bestMetrics.Score
		fmt.Printf("site: %s connections: %d queue: %d score: %d\n", bestMetrics.Site, bestMetrics.Connections, bestMetrics.Queue, bestMetrics.Score)
	} else {
		bestScore = ConnectionMax*Weight + QueueMax + 1000 // High fallback score
	}

	metricsChan := make(chan Metrics, len(sites))
	activeGoroutines := 0

	for _, site := range sites {
		if site == bestSite || site == "" {
			continue
		}

		activeGoroutines++
		go func(s string) {
			metricsChan <- m.getMetrics(s)
		}(site)
	}

	for i := 0; i < activeGoroutines; i++ {
		metrics := <-metricsChan
		fmt.Printf("site: %s connections: %d queue: %d score: %d\n", metrics.Site, metrics.Connections, metrics.Queue, metrics.Score)
		if metrics.Error == nil && metrics.Score < bestScore {
			bestSite = metrics.Site
			bestScore = metrics.Score
		}
	}

	m.Site = bestSite

	return bestSite, nil
}

func main() {
	var err error

	kingpin.MustParse(app.Parse(os.Args[1:]))

	m := NewMonitor()

	m.Host, err = m.getLocal()
	if err != nil {
		os.Exit(1)
	}

	site, err := m.runTest(strings.Split(*sites, ","))
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("best site: %s\n", site)
}

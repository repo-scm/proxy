package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

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

func (m *Monitor) calculateScore(connections, queue int) int {
	return connections*Weight + queue
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
			Score:       m.calculateScore(ConnectionMax, QueueMax),
			Error:       fmt.Errorf("failed to get metrics for site %s", site),
		}
	}

	score := m.calculateScore(res.connections, res.queue)

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
	} else {
		bestScore = ConnectionMax*Weight + QueueMax
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

	fmt.Println(site)
}

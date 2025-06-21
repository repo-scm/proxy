package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	locationChengdu  = "chengdu"
	locationShanghai = "shanghai"
	locationXian     = "xian"

	hostPrefix = "10."
	siteName   = "gerrit"

	ConnectionMax = 65536
	QueueMax      = 65536
	Weights       = 10
)

var (
	Sites = []string{
		"192.168.0.1",
		"192.168.0.2",
	}
)

type Monitor struct {
	Host string
	Site string
}

func NewMonitor() *Monitor {
	return &Monitor{}
}

func (m *Monitor) getHost() (string, error) {
	var host string

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Wrap(err, "failed to get host\n")
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip := ipnet.IP.String()
				if strings.HasPrefix(ip, hostPrefix) {
					host = ip
					break
				}
			}
		}
	}

	if host == "" {
		return "", errors.New("failed to get host\n")
	}

	return host, nil
}

func (m *Monitor) getEnv() error {
	return nil
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
					fmt.Printf("%s GERRIT_QUEUE: %d\n", host, queue)
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
						fmt.Printf("%s GERRIT_CONNECTION: %d\n", host, connection)
						return connection, nil
					}
				}
			}
		}
	}

	return ConnectionMax, nil
}

func (m *Monitor) runTest(sites []string) error {
	for _, site := range sites {
		if site == m.Site {
			continue
		}

		connectionSelected, _ := m.getConnection(m.Site)
		queueSelected, _ := m.getQueue(m.Site)
		valueSelected := connectionSelected*Weights + queueSelected

		connectionItem, _ := m.getConnection(site)
		queueItem, _ := m.getQueue(site)
		valueItem := connectionItem*Weights + queueItem

		if valueItem < valueSelected {
			m.Site = site
		}
	}

	return nil
}

func (m *Monitor) testLocationChengdu() error {
	return nil
}

func (m *Monitor) testLocationShanghai() error {
	locations := strings.Join(Sites, ",")
	fmt.Printf("GERRIT_SH=%s\n", locations)

	servers := strings.Split(locations, ",")
	m.Site = servers[0]

	return m.runTest(servers)
}

func (m *Monitor) testLocationXian() error {
	return nil
}

func (m *Monitor) testLocation(name string) error {
	var err error

	switch name {
	case locationChengdu:
		err = m.testLocationChengdu()
	case locationShanghai:
		err = m.testLocationShanghai()
	case locationXian:
		err = m.testLocationXian()
	default:
		err = m.testLocationShanghai()
	}

	return err
}

func (m *Monitor) chooseLocation() (string, error) {
	return locationShanghai, nil
}

func main() {
	var err error

	m := NewMonitor()

	m.Host, err = m.getHost()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := m.getEnv(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	name, _ := m.chooseLocation()
	if err := m.testLocation(name); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	_ = os.Setenv("LOCAL_GERRIT", m.Site)
	fmt.Printf("LOCAL_GERRIT=%s\n", m.Site)
}

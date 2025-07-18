package monitor

import (
	"time"
)

func GetTestSitesData() []*SiteStatus {
	now := time.Now()

	return []*SiteStatus{
		{
			Name:         "gerrit-beijing",
			Location:     "Beijing, China",
			Url:          "https://gerrit-beijing.com",
			Host:         "10.67.16.29",
			Healthy:      true,
			ResponseTime: 45,
			Connections:  3,
			QueueSize:    2,
			Score:        47,
			LastCheck:    now.Add(-time.Minute * 2),
			Error:        "",
		},
		{
			Name:         "gerrit-shanghai",
			Location:     "Shanghai, China",
			Url:          "https://gerrit-shanghai.com",
			Host:         "10.63.237.206",
			Healthy:      true,
			ResponseTime: 52,
			Connections:  1,
			QueueSize:    0,
			Score:        15,
			LastCheck:    now.Add(-time.Minute * 1),
			Error:        "",
		},
		{
			Name:         "gerrit-chengdu",
			Location:     "Chengdu, China",
			Url:          "https://gerrit-chengdu.com",
			Host:         "10.75.200.210",
			Healthy:      true,
			ResponseTime: 38,
			Connections:  5,
			QueueSize:    1,
			Score:        55,
			LastCheck:    now.Add(-time.Minute * 3),
			Error:        "",
		},
		{
			Name:         "gerrit-xian",
			Location:     "Xi'an, China",
			Url:          "https://gerrit-xian.com",
			Host:         "10.95.243.159",
			Healthy:      false,
			ResponseTime: -1,
			Connections:  ConnectionMax,
			QueueSize:    QueueMax,
			Score:        -1,
			LastCheck:    now.Add(-time.Minute * 5),
			Error:        "Connection timeout - site unreachable",
		},
	}
}

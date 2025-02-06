package collector

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	probing "github.com/prometheus-community/pro-bing"
	log "github.com/sirupsen/logrus"
	"github.com/spanwalla/docker-monitoring-pinger/internal/entity"
	"time"
)

type DockerCollector struct {
	client *client.Client
}

func NewDockerCollector() (*DockerCollector, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("NewDockerCollector - NewClientWithOpts: %v", err)
	}
	return &DockerCollector{
		client: apiClient,
	}, nil
}

func (d *DockerCollector) Collect() ([]entity.Report, error) {
	ctx := context.Background()
	var err error

	var containers []types.Container
	containers, err = d.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("getContainerList - ContainerList: %v", err)
	}

	result := make([]entity.Report, 0, len(containers))
	for _, cnt := range containers {
		// Inspect the container
		var inspect types.ContainerJSON
		inspect, err = d.client.ContainerInspect(ctx, cnt.ID)
		if err != nil {
			return nil, fmt.Errorf("getContainerInspect - ContainerInspect: %v", err)
		}

		// Get IP from Network Settings
		ipAddress := ""
		if inspect.NetworkSettings != nil {
			for _, network := range inspect.NetworkSettings.Networks {
				if network.IPAddress != "" {
					ipAddress = network.IPAddress
					break
				}
			}
		}

		pingLatency := -1
		timestamp := time.Now()
		if ipAddress != "" {
			var latency int
			latency, err = d.pingContainer(ipAddress)
			if err != nil {
				log.Errorf("DockerCollector.Collect: %v", err)
			} else {
				pingLatency = latency
			}
		}

		result = append(result, entity.Report{
			ContainerId: cnt.ID,
			Ip:          ipAddress,
			Latency:     pingLatency,
			Status:      cnt.Status,
			State:       cnt.State,
			Timestamp:   timestamp,
		})
	}
	return result, nil
}

func (d *DockerCollector) pingContainer(ipAddress string) (int, error) {
	pinger, err := probing.NewPinger(ipAddress)
	if err != nil {
		return 0, fmt.Errorf("pingContainer - NewPinger: %v", err)
	}

	pinger.Count = 3
	pinger.Timeout = 5 * time.Second
	pinger.SetPrivileged(false)

	err = pinger.Run()
	if err != nil {
		return 0, fmt.Errorf("pingContainer - Ping %s: %v", ipAddress, err)
	}

	stats := pinger.Statistics()
	if stats.PacketLoss == 100 {
		return -1, nil
	}

	return int(stats.AvgRtt.Milliseconds()), nil
}

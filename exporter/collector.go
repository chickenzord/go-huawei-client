package main

import (
	"fmt"

	"github.com/chickenzord/go-huawei-client/pkg/eg8145v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type RouterCollector struct {
	client *eg8145v5.Client

	deviceOnline  *prometheus.GaugeVec
	resourceUsage *prometheus.GaugeVec
}

func NewRouterCollector(cfg *eg8145v5.Config) *RouterCollector {
	return &RouterCollector{
		client: eg8145v5.NewClient(*cfg),
		deviceOnline: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "router",
			Name:      "device_online",
			Help:      "Device online status in the router",
		}, []string{
			"mac_address", "hostname",
		}),
		resourceUsage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "router",
			Name:      "resource_usage",
			Help:      "Router resource usages",
		}, []string{
			"type", "unit",
		}),
	}
}

func (c *RouterCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c.deviceOnline, ch)
	prometheus.DescribeByCollect(c.resourceUsage, ch)
}

func (c *RouterCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.client.Session(func(client *eg8145v5.Client) error {
		devices, err := client.ListUserDevices()
		if err != nil {
			return fmt.Errorf("failed to list user devices: %w", err)
		}

		for _, device := range devices {
			deviceOnline := c.deviceOnline.With(prometheus.Labels{
				"mac_address": device.MacAddr,
				"hostname":    device.HostName,
			})

			if device.Online() {
				deviceOnline.Set(1)
			} else {
				deviceOnline.Set(0)
			}

			ch <- deviceOnline
		}

		usage, err := client.GetResourceUsage()
		if err != nil {
			return fmt.Errorf("failed to get resource usage: %w", err)
		}

		memoryUsage := c.resourceUsage.With(prometheus.Labels{
			"type": "memory",
			"unit": "percent",
		})
		memoryUsage.Set(float64(usage.Memory))
		ch <- memoryUsage

		cpuUsage := c.resourceUsage.With(prometheus.Labels{
			"type": "cpu",
			"unit": "percent",
		})
		cpuUsage.Set(float64(usage.CPU))
		ch <- cpuUsage

		return nil
	}); err != nil {
		log.Err(err).Msg("error collecting router metrics")
	}
}

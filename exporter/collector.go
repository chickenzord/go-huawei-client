package main

import (
	"fmt"

	"github.com/chickenzord/go-huawei-client/pkg/hn8010ts"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type RouterCollector struct {
	client *hn8010ts.Client

	levels *prometheus.GaugeVec
}

func NewRouterCollector(cfg *hn8010ts.Config) *RouterCollector {
	return &RouterCollector{
		client: hn8010ts.NewClient(*cfg),
		levels: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "huawei_ont_optic",
			Name:      "level",
			Help:      "Optical power",
		}, []string{
			"direction",
		}),
	}
}

func (c *RouterCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c.levels, ch)
}

func (c *RouterCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.client.Session(func(client *hn8010ts.Client) error {
		opticInfo, err := client.GetOpticInfo()
		if err != nil {
			return fmt.Errorf("failed to list user devices: %w", err)
		}

		rx := c.levels.WithLabelValues("rx")
		rx.Set(float64(opticInfo.RXPower))
		ch <- rx

		tx := c.levels.WithLabelValues("tx")
		tx.Set(float64(opticInfo.TXPower))
		ch <- tx

		return nil
	}); err != nil {
		log.Err(err).Msg("error collecting router metrics")
	}
}

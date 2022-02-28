package collector

import (
	"context"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type StatisticsCollector struct {
	Cache        *cache.Cache
	Logger       *zap.SugaredLogger
	ErrorCounter *Counter
}

func (c *StatisticsCollector) Run(ctx context.Context) error {
	return nil
}

func (c *StatisticsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *StatisticsCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"vmanage_exporter_scrape_errors",
			"Number of scrape errors",
			[]string{},
			nil,
		),
		prometheus.GaugeValue,
		float64(c.ErrorCounter.Get()),
	)
}

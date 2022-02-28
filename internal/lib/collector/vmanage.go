package collector

import (
	"context"
	"fmt"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zebbra/vmanage-exporter/internal/lib/vmanage"
	"go.uber.org/zap"
	"sync"
	"time"
)

type VmanageCollector struct {
	Cache         *cache.Cache
	Client        *vmanage.Client
	Logger        *zap.SugaredLogger
	ErrorCounter  *Counter
	ScrapeCounter *Counter
}

func (c *VmanageCollector) Run(ctx context.Context) error {
	c.Logger.Infow("Refresh device list")
	startTime := time.Now()

	devices := map[string]vmanage.Device{}
	var queue chan string

	if devs, err := c.Client.Device(ctx); err == nil {
		queue = make(chan string, len(devs))

		for _, d := range devs {
			devices[d.DeviceID] = d
			queue <- d.DeviceID
		}

		c.Logger.Infow("Successfully refreshed device list", "count", len(devices))
		close(queue)

	} else {
		c.Logger.Errorw(
			"Error fetching device list",
			"error", err,
		)

		c.ErrorCounter.Inc()
		return err
	}

	synced := true

	c.Cache.Set("devices", devices, cache.DefaultExpiration)

	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()

		for deviceID := range queue {
			select {
			case <-ctx.Done():
				c.Logger.Warnw(
					"Timed out refreshing interface statistics",
					"error", ctx.Err(),
				)
				continue

			default:
				// fetch interface statistics
				{
					c.Logger.Infow("Refresh interface statistics", "DeviceID", deviceID)

					res, err := c.Client.DeviceInterface(
						ctx,
						synced,
						&vmanage.DeviceInterfaceListOptions{DeviceID: deviceID},
					)

					if err != nil {
						c.Logger.Warnw(
							"Error fetching interface statistics",
							"DeviceID", deviceID,
							"error", err,
						)

						c.ErrorCounter.Inc()
					}

					c.Cache.Set(fmt.Sprintf("ifs_%s", deviceID), res, cache.DefaultExpiration)
				}

				{
					c.Logger.Infow("Refresh system statistics", "DeviceID", deviceID)

					res, err := c.Client.DeviceSystemStatus(
						ctx,
						false, // TODO: synced did not return valid data in test env?!
						&vmanage.DeviceSystemStatusListOptions{DeviceID: deviceID},
					)

					if err != nil {
						c.Logger.Warnw(
							"Error fetching system statistics",
							"DeviceID", deviceID,
							"error", err,
						)

						c.ErrorCounter.Inc()

						continue
					}

					if len(res) != 1 {
						c.Logger.Warnw(
							"Error fetching system statistics: should return a single entry",
							"DeviceID", deviceID,
							"data", res,
						)

						c.ErrorCounter.Inc()

						continue
					}

					c.Cache.Set(fmt.Sprintf("system_status_%s", deviceID), res[0], cache.DefaultExpiration)
				}

			}
		}
	}

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go worker()
	}

	wg.Wait()
	c.Logger.Infow(
		"Refresh done",
		"duration",
		time.Now().Sub(startTime),
	)

	return nil
}

func (c *VmanageCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *VmanageCollector) Collect(ch chan<- prometheus.Metric) {
	devices := map[string]vmanage.Device{}

	if d, found := c.Cache.Get("devices"); found {
		devices = d.(map[string]vmanage.Device)
	}

	// general stats
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			"vmanage_devices",
			"Number of devices managed by vmanage",
			[]string{},
			nil,
		),
		prometheus.GaugeValue,
		float64(len(devices)),
	)

	status := func(s string) float64 {
		if s == "normal" {
			return 1
		}

		return 0
	}

	reachable := func(b bool) float64 {
		if b {
			return 1
		}

		return 0
	}

	uptime := func(ts int64) float64 {
		return float64(time.Now().UnixMilli() - ts)
	}

	for _, d := range devices {
		deviceLabels := deviceLabels(d)

		// device info
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"vmanage_device_info",
				"Info about device",
				deviceLabelsInfo(d).Labels,
				nil,
			),
			prometheus.GaugeValue,
			status(d.Status),
			deviceLabelsInfo(d).Values...,
		)

		// device stats
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"vmanage_device_status",
				"Status of device",
				append(deviceLabels.Labels, "status"),
				nil,
			),
			prometheus.GaugeValue,
			status(d.Status),
			append(deviceLabels.Values, d.Status)...,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"vmanage_device_reachability",
				"Reachability of device",
				append(deviceLabels.Labels, "reachability"),
				nil,
			),
			prometheus.GaugeValue,
			reachable(d.IsReachable()),
			append(deviceLabels.Values, d.Reachability)...,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc(
				"vmanage_device_uptime",
				"Uptime of device",
				deviceLabels.Labels,
				nil,
			),
			prometheus.CounterValue,
			uptime(d.UptimeDate),
			deviceLabels.Values...,
		)

		// system stats
		if ss, found := c.Cache.Get(fmt.Sprintf("system_status_%s", d.DeviceID)); found {
			ss := ss.(vmanage.DeviceSystemStatus)
			mem := ss.Memory()
			cpu := ss.CPU()

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_mem_used",
					"Memory Used",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				float64(mem.Used),
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_mem_free",
					"Memory Free",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				float64(mem.Free),
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_mem_total",
					"Memory Total",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				float64(mem.Total),
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_cpu_user_percentage",
					"CPU User(%)",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				cpu.UserPercentage,
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_cpu_system_percentage",
					"CPU System(%)",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				cpu.SystemPercentage,
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_cpu_idle_percentage",
					"CPU Idle(%)",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				cpu.IdlePercentage,
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_load_avg1",
					"Load Average 1 min",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				cpu.LoadAvg1,
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_load_avg5",
					"Load Average 5 min",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				cpu.LoadAvg5,
				deviceLabels.Values...,
			)

			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					"vmanage_device_load_avg15",
					"Load Average 15 min",
					deviceLabels.Labels,
					nil,
				),
				prometheus.GaugeValue,
				cpu.LoadAvg15,
				deviceLabels.Values...,
			)
		}

		// interface stats
		if ifs, found := c.Cache.Get(fmt.Sprintf("ifs_%s", d.DeviceID)); found {
			for _, i := range ifs.([]vmanage.DeviceInterface) {
				ifLabels := interfaceLabels(d, i)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_tx_octets",
						"Interface TX Octets",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.TxOctets),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_rx_octets",
						"Interface RX Octets",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.RxOctets),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_tx_packets",
						"Interface TX Unicast Packets",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.TxPackets),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_rx_packets",
						"Interface RX Unicast Packets",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.RxPackets),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_tx_errors",
						"Interface Tx Errors",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.TxErrors),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_rx_errors",
						"Interface Rx Errors",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.RxErrors),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_tx_drops",
						"Interface Tx Drops",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.TxDrops),
					ifLabels.Values...,
				)

				ch <- prometheus.MustNewConstMetric(
					prometheus.NewDesc(
						"vmanage_device_interface_rx_drops",
						"Interface Rx Drops",
						ifLabels.Labels,
						nil,
					),
					prometheus.CounterValue,
					float64(i.RxDrops),
					ifLabels.Values...,
				)
			}
		}
	}

}

func deviceLabelsInfo(d vmanage.Device) struct {
	Labels []string
	Values []string
} {
	l := []string{"DeviceID", "SystemIP", "Hostname", "DeviceModel", "Version", "DeviceOS"}
	v := []string{
		d.DeviceID,
		d.SystemIP,
		d.Hostname,
		d.DeviceModel,
		d.Version,
		d.DeviceOS,
	}

	return struct {
		Labels []string
		Values []string
	}{
		Labels: l,
		Values: v,
	}
}

func deviceLabels(d vmanage.Device) struct {
	Labels []string
	Values []string
} {
	l := []string{"DeviceID", "Hostname"}
	v := []string{
		d.DeviceID,
		d.Hostname,
	}

	return struct {
		Labels []string
		Values []string
	}{
		Labels: l,
		Values: v,
	}
}

func interfaceLabels(d vmanage.Device, i vmanage.DeviceInterface) struct {
	Labels []string
	Values []string
} {
	l := []string{"DeviceID", "VdeviceName", "Ifname", "IfIndex", "AfType", "VdeviceDataKey"}
	v := []string{
		d.DeviceID,
		i.VdeviceName,
		i.Ifname,
		fmt.Sprintf("%d", i.IfIndexInt()),
		i.AfType,
		i.VdeviceDataKey,
	}

	return struct {
		Labels []string
		Values []string
	}{
		Labels: l,
		Values: v,
	}
}

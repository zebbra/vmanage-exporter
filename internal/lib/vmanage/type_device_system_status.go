package vmanage

import (
	"context"
	"github.com/google/go-querystring/query"
	"net/url"
	"strconv"
)

func (c *Client) DeviceSystemStatus(ctx context.Context, synced bool, options *DeviceSystemStatusListOptions) ([]DeviceSystemStatus, error) {
	endpoint := "/dataservice/device/system/status"

	if synced {
		endpoint = "/dataservice/device/system/synced/status"
	}

	resp, err := c.Fetch(
		ctx,
		endpoint,
		options,
		&DeviceSystemStatusList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*DeviceSystemStatusList)
	return list.Data, nil
}

type DeviceSystemStatus struct {
	MemUsed                  string `json:"mem_used"`
	Procs                    string `json:"procs"`
	DiskAvail                string `json:"disk_avail"`
	DiskMount                string `json:"disk_mount"`
	BoardType                string `json:"board_type"`
	VdeviceName              string `json:"vdevice-name"`
	TotalCPUCount            string `json:"total_cpu_count"`
	MemCached                string `json:"mem_cached"`
	Timezone                 string `json:"timezone"`
	DiskFs                   string `json:"disk_fs"`
	FpCPUCount               string `json:"fp_cpu_count"`
	ChassisSerialNumber      string `json:"chassis-serial-number"`
	Min1Avg                  string `json:"min1_avg"`
	StateDescription         string `json:"state_description"`
	Personality              string `json:"personality"`
	DiskUsed                 string `json:"disk_used"`
	DiskUse                  string `json:"disk_use"`
	DiskStatus               string `json:"disk_status"`
	State                    string `json:"state"`
	ConfigDateDateTimeString string `json:"config_date/date-time-string"`
	LinuxCPUCount            string `json:"linux_cpu_count"`
	CPUUser                  string `json:"cpu_user"`
	TestbedMode              string `json:"testbed_mode"`
	Min15Avg                 string `json:"min15_avg"`
	DiskSize                 string `json:"disk_size"`
	CPUIdle                  string `json:"cpu_idle"`
	MemBuffers               string `json:"mem_buffers"`
	ModelSku                 string `json:"model_sku"`
	CPUSystem                string `json:"cpu_system"`
	Version                  string `json:"version"`
	Min5Avg                  string `json:"min5_avg"`
	TcpdCPUCount             string `json:"tcpd_cpu_count"`
	VdeviceHostName          string `json:"vdevice-host-name"`
	MemTotal                 string `json:"mem_total"`
	Uptime                   string `json:"uptime"`
	VdeviceDataKey           string `json:"vdevice-dataKey"`
	MemFree                  string `json:"mem_free"`
	BootloaderVersion        string `json:"bootloader_version"`
	FipsMode                 string `json:"fips_mode"`
	BuildNumber              string `json:"build_number"`
	Lastupdated              int64  `json:"lastupdated"`
	LoghostStatus            string `json:"loghost_status"`
	UptimeDate               int64  `json:"uptime-date"`
}

func (d *DeviceSystemStatus) Memory() DeviceSystemStatusMemory {
	u, _ := strconv.Atoi(d.MemUsed)
	f, _ := strconv.Atoi(d.MemFree)
	t, _ := strconv.Atoi(d.MemTotal)
	b, _ := strconv.Atoi(d.MemBuffers)
	c, _ := strconv.Atoi(d.MemCached)

	return DeviceSystemStatusMemory{
		Used:    u,
		Free:    f,
		Total:   t,
		Buffers: b,
		Cached:  c,
	}
}

type DeviceSystemStatusMemory struct {
	Used    int
	Free    int
	Total   int
	Buffers int
	Cached  int
}

func (d *DeviceSystemStatus) CPU() DeviceSystemStatusCPU {
	u, _ := strconv.ParseFloat(d.CPUUser, 64)
	s, _ := strconv.ParseFloat(d.CPUSystem, 64)
	i, _ := strconv.ParseFloat(d.CPUIdle, 64)
	l1, _ := strconv.ParseFloat(d.Min1Avg, 64)
	l5, _ := strconv.ParseFloat(d.Min5Avg, 64)
	l15, _ := strconv.ParseFloat(d.Min15Avg, 64)

	return DeviceSystemStatusCPU{
		UserPercentage:   u,
		SystemPercentage: s,
		IdlePercentage:   i,
		LoadAvg1:         l1,
		LoadAvg5:         l5,
		LoadAvg15:        l15,
	}
}

type DeviceSystemStatusCPU struct {
	UserPercentage   float64
	SystemPercentage float64
	IdlePercentage   float64
	LoadAvg1         float64
	LoadAvg5         float64
	LoadAvg15        float64
}

type DeviceSystemStatusList struct {
	Data []DeviceSystemStatus `json:"data"`
}

type DeviceSystemStatusListOptions struct {
	DeviceID string `url:"deviceId,omitempty"`
}

func (o *DeviceSystemStatusListOptions) Params() url.Values {
	v, _ := query.Values(o)
	return v
}

package vmanage

import "context"

func (c *Client) DeviceStateSystemStatus(ctx context.Context) ([]DeviceStateSystemStatus, error) {
	resp, err := c.Fetch(
		ctx,
		"/dataservice/data/device/state/SystemStatus",
		nil,
		&DeviceStateSystemStatusList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*DeviceStateSystemStatusList)
	return list.Data, nil
}

type DeviceStateSystemStatus struct {
	RecordID                string `json:"recordId"`
	BoardType               string `json:"board_type"`
	VdeviceName             string `json:"vdevice-name"`
	TotalCPUCount           string `json:"total_cpu_count"`
	RebootType              string `json:"reboot_type"`
	FpCPUCount              string `json:"fp_cpu_count,omitempty"`
	StateDescription        string `json:"state_description"`
	Rid                     int    `json:"@rid"`
	Personality             string `json:"personality"`
	DiskStatus              string `json:"disk_status"`
	State                   string `json:"state"`
	LinuxCPUCount           string `json:"linux_cpu_count,omitempty"`
	RebootReason            string `json:"reboot_reason"`
	TestbedMode             string `json:"testbed_mode"`
	CreateTimeStamp         int64  `json:"createTimeStamp"`
	ModelSku                string `json:"model_sku"`
	VdeviceHostName         string `json:"vdevice-host-name"`
	Version                 string `json:"version"`
	TcpdCPUCount            string `json:"tcpd_cpu_count,omitempty"`
	VdeviceDataKey          string `json:"vdevice-dataKey"`
	VmanageSystemIP         string `json:"vmanage-system-ip"`
	BootloaderVersion       string `json:"bootloader_version"`
	FipsMode                string `json:"fips_mode"`
	Lastupdated             int64  `json:"lastupdated"`
	BuildNumber             string `json:"build_number"`
	LoghostStatus           string `json:"loghost_status"`
	UptimeDate              int64  `json:"uptime-date"`
	VmanageStorageDiskMount string `json:"vmanage-storage-disk-mount,omitempty"`
}

type DeviceStateSystemStatusList struct {
	Data []DeviceStateSystemStatus `json:"data"`
}

package vmanage

import "context"

func (c *Client) Device(ctx context.Context) ([]Device, error) {
	resp, err := c.Fetch(
		ctx,
		"/dataservice/device",
		nil,
		&DeviceList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*DeviceList)
	return list.Data, nil
}

type Device struct {
	DeviceID            string   `json:"deviceId"`
	SystemIP            string   `json:"system-ip"`
	Hostname            string   `json:"host-name"`
	Reachability        string   `json:"reachability"`
	Status              string   `json:"status"`
	Personality         string   `json:"personality"`
	DeviceType          string   `json:"device-type"`
	Timezone            string   `json:"timezone"`
	DeviceGroups        []string `json:"device-groups"`
	Lastupdated         int64    `json:"lastupdated"`
	DomainID            string   `json:"domain-id,omitempty"`
	BoardSerial         string   `json:"board-serial"`
	CertificateValidity string   `json:"certificate-validity"`
	MaxControllers      string   `json:"max-controllers,omitempty"`
	UUID                string   `json:"uuid"`
	ControlConnections  string   `json:"controlConnections,omitempty"`
	DeviceModel         string   `json:"device-model"`
	Version             string   `json:"version"`
	ConnectedVManages   []string `json:"connectedVManages"`
	SiteID              string   `json:"site-id"`
	Latitude            string   `json:"latitude"`
	Longitude           string   `json:"longitude"`
	IsDeviceGeoData     bool     `json:"isDeviceGeoData"`
	Platform            string   `json:"platform"`
	UptimeDate          int64    `json:"uptime-date"`
	StatusOrder         int      `json:"statusOrder"`
	DeviceOS            string   `json:"device-os"`
	Validity            string   `json:"validity"`
	State               string   `json:"state"`
	StateDescription    string   `json:"state_description"`
	ModelSKU            string   `json:"model_sku"`
	LocalSystemIP       string   `json:"local-system-ip"`
	TotalCPUCount       string   `json:"total_cpu_count"`
	TestbedMode         bool     `json:"testbed_mode"`
	LayoutLevel         int      `json:"layoutLevel"`
	OmpPeers            string   `json:"ompPeers,omitempty"`
	LinuxCPUCount       string   `json:"linux_cpu_count,omitempty"`
}

func (d *Device) IsReachable() bool {
	if d.Reachability == "reachable" {
		return true
	}

	return false
}

type DeviceList struct {
	Data []Device `json:"data"`
}

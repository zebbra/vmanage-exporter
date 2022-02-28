package vmanage

import "context"

func (c *Client) DeviceMonitor(ctx context.Context) ([]DeviceMonitor, error) {
	resp, err := c.Fetch(
		ctx,
		"/dataservice/device/monitor",
		nil,
		&DeviceMonitorList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*DeviceMonitorList)
	return list.Data, nil
}

type DeviceMonitor struct {
	DeviceModel string `json:"device-model"`
	DeviceType  string `json:"device-type"`
	SystemIP    string `json:"system-ip"`
	ID          int    `json:"_id"`
	HostName    string `json:"host-name"`
	SiteID      string `json:"site-id"`
	LayoutLevel int    `json:"layoutLevel"`
	Status      string `json:"status"`
}

type DeviceMonitorList struct {
	Data []DeviceMonitor `json:"data"`
}

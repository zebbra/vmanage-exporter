package vmanage

import "context"

func (c *Client) DeviceCounter(ctx context.Context) ([]DeviceCounter, error) {
	resp, err := c.Fetch(
		ctx,
		"/dataservice/device/counters",
		nil,
		&DeviceCounterList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*DeviceCounterList)
	return list.Data, nil
}

type DeviceCounter struct {
	SystemIP                       string `json:"system-ip"`
	NumberVsmartControlConnections int    `json:"number-vsmart-control-connections,omitempty"`
	ExpectedControlConnections     int    `json:"expectedControlConnections,omitempty"`
	OmpPeersUp                     int    `json:"ompPeersUp,omitempty"`
	OmpPeersDown                   int    `json:"ompPeersDown,omitempty"`
	RebootCount                    int    `json:"rebootCount"`
	CrashCount                     int    `json:"crashCount"`
}

type DeviceCounterList struct {
	Data []DeviceCounter `json:"data"`
}

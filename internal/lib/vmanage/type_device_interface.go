package vmanage

import (
	"context"
	"github.com/google/go-querystring/query"
	"net/url"
	"strconv"
)

func (c *Client) DeviceInterface(ctx context.Context, synced bool, options *DeviceInterfaceListOptions) ([]DeviceInterface, error) {
	endpoint := "/dataservice/device/interface"

	if synced {
		endpoint = "/dataservice/device/interface/synced"
	}

	resp, err := c.Fetch(
		ctx,
		endpoint,
		options,
		&DeviceInterfaceList{},
	)

	if err != nil {
		return nil, err
	}

	list := resp.(*DeviceInterfaceList)
	return list.Data, nil
}

// TODO: consolidate different interface types?
type DeviceInterface struct {
	VdeviceName      string      `json:"vdevice-name"`
	RxErrors         int         `json:"rx-errors,omitempty"`
	TxKbps           int         `json:"tx-kbps,omitempty"`
	IfAdminStatus    string      `json:"if-admin-status"`
	TxErrors         int         `json:"tx-errors,omitempty"`
	TxPps            int         `json:"tx-pps,omitempty"`
	Ifname           string      `json:"ifname"`
	RxPps            int         `json:"rx-pps,omitempty"`
	AfType           string      `json:"af-type"`
	IfOperStatus     string      `json:"if-oper-status"`
	IfIndex          interface{} `json:"ifindex"` // some types return int, some string
	RxPackets        int         `json:"rx-packets,omitempty"`
	SecondaryAddress string      `json:"secondary-address,omitempty"`
	VpnID            string      `json:"vpn-id"`
	VdeviceHostName  string      `json:"vdevice-host-name"`
	RxDrops          int         `json:"rx-drops,omitempty"`
	TxDrops          int         `json:"tx-drops,omitempty"`
	Uptime           string      `json:"uptime,omitempty"`
	Ipv6Address      string      `json:"ipv6-address"`
	Secondary        string      `json:"secondary,omitempty"`
	IPAddress        string      `json:"ip-address"`
	Hwaddr           string      `json:"hwaddr,omitempty"`
	VdeviceDataKey   string      `json:"vdevice-dataKey"`
	TxOctets         int         `json:"tx-octets,omitempty"`
	TxPackets        int         `json:"tx-packets,omitempty"`
	RxOctets         int         `json:"rx-octets,omitempty"`
	RxKbps           int         `json:"rx-kbps,omitempty"`
	Lastupdated      int64       `json:"lastupdated"`
	PortType         string      `json:"port-type,omitempty"`
	UptimeDate       int64       `json:"uptime-date,omitempty"`
	EncapType        string      `json:"encap-type,omitempty"`
}

func (d *DeviceInterface) IfIndexInt() int {
	switch v := d.IfIndex.(type) {
	case int:
		return v
	case string:
		i, _ := strconv.Atoi(v)
		return i
	case float64:
		return int(v)
	default:
		return 0
	}
}

func (d *DeviceInterface) IsUpAdmin() bool {
	switch d.IfAdminStatus {
	case "Up":
		return true
	case "if-state-up":
		return true
	default:
		return false
	}
}

func (d *DeviceInterface) IsUpOper() bool {
	switch d.IfOperStatus {
	case "Up":
		return true
	case "if-state-up":
		return true
	default:
		return false
	}
}

type DeviceInterfaceList struct {
	Data []DeviceInterface `json:"data"`
}

type DeviceInterfaceListOptions struct {
	VpnID    string `url:"vpn-id,omitempty"`
	Ifname   string `url:"ifname,omitempty"`
	AfType   string `url:"af-type,omitempty"`
	DeviceID string `url:"deviceId,omitempty"`
}

func (o *DeviceInterfaceListOptions) Params() url.Values {
	v, _ := query.Values(o)
	return v
}

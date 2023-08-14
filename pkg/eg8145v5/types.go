package eg8145v5

import "strings"

type UserDevice struct {
	Domain                  string
	IpAddr                  string
	MacAddr                 string
	Port                    string
	PortID                  string
	DevStatus               string
	IpType                  string
	Time                    string
	HostName                string
	IPv4Enabled             string
	IPv6Enabled             string
	DeviceType              string
	UserDevAlias            string
	UserSpecifiedDeviceType string
	LeaseTimeRemaining      string
}

func (d *UserDevice) Online() bool {
	return strings.EqualFold(d.DevStatus, "online")
}

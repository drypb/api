package proxmox

import (
	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// NetworkIPAddress represents an IP address associated with a network
// interface.
type NetworkIPAddress struct {
	IPAddressType string // "ipv4" or "ipv6"
	IPAddress     string
	Prefix        int
	MacAddress    string
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	Name            string
	HardwareAddress string
	IPAddresses     []*NetworkIPAddress
}

// fromAPIIPAddress converts a proxmoxAPI.AgentNetworkIPAddress to a
// NetworkIPAddress.
func fromAPIIPAddress(apiIP *proxmoxAPI.AgentNetworkIPAddress) *NetworkIPAddress {
	return &NetworkIPAddress{
		IPAddressType: apiIP.IPAddressType,
		IPAddress:     apiIP.IPAddress,
		Prefix:        apiIP.Prefix,
		MacAddress:    apiIP.MacAddress,
	}
}

// fromAPIInterface converts a proxmoxAPI.AgentNetworkIface to a
// NetworkInterface.
func fromAPIInterface(apiIface *proxmoxAPI.AgentNetworkIface) *NetworkInterface {
	var ipAddresses []*NetworkIPAddress
	for _, apiIP := range apiIface.IPAddresses {
		ipAddresses = append(ipAddresses, fromAPIIPAddress(apiIP))
	}
	return &NetworkInterface{
		Name:            apiIface.Name,
		HardwareAddress: apiIface.HardwareAddress,
		IPAddresses:     ipAddresses,
	}
}

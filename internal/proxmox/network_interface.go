package proxmox

import (
	proxmoxAPI "github.com/luthermonson/go-proxmox"
)

// NetworkIPAddress represents an IP address.
type NetworkIPAddress struct {
	IPAddressType string // IPAddressType is "ipv4" or "ipv6".
	IPAddress     string // IPAddress is the netid+hostid.
	Prefix        int    // Prefix is the CIDR.
	MacAddress    string // MacAddress is the MAC address.
}

// NetworkInterface represents a network interface.
type NetworkInterface struct {
	Name            string
	HardwareAddress string
	IPAddresses     []*NetworkIPAddress
}

// fromAPIIPAddress converts a [proxmoxAPI.AgentNetworkIPAddress] to a
// [NetworkIPAddress].
func fromAPIIPAddress(apiIP *proxmoxAPI.AgentNetworkIPAddress) *NetworkIPAddress {
	return &NetworkIPAddress{
		IPAddressType: apiIP.IPAddressType,
		IPAddress:     apiIP.IPAddress,
		Prefix:        apiIP.Prefix,
		MacAddress:    apiIP.MacAddress,
	}
}

// fromAPIInterface converts a [proxmoxAPI.AgentNetworkIface] to a
// [NetworkInterface].
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

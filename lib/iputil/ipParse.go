package iputil

import (
	"errors"
	"github.com/apparentlymart/go-cidr/cidr"
	"net"
	"strconv"
)

// CheckNetwork : Get IP address and netmask as string value then check if valid.
// Return network as *net.IPNet if valid.
func CheckNetwork(IP string, networkNetmask string) (*net.IPNet, error) {
	netIP := CheckValidIP(IP)
	if netIP == nil {
		return nil, errors.New("invalid IP address")
	}

	mask, err := CheckNetmask(networkNetmask)
	if err != nil {
		return nil, err
	}

	maskLen, _ := mask.Size()
	_, netNetwork, err := net.ParseCIDR(IP + "/" + strconv.Itoa(maskLen))
	if err != nil {
		return nil, err
	}

	return netNetwork, nil
}

// GetFirstAndLastIPs : Return first and last IP addresses from given network IP address and netmask.
// Return as net.IP for both first and last IP addresses.
func GetFirstAndLastIPs(networkIP string, networkNetmask string) (net.IP, net.IP, error) {
	netNetwork, err := CheckNetwork(networkIP, networkNetmask)
	if err != nil {
		return nil, nil, err
	}

	firstIP, lastIP := cidr.AddressRange(netNetwork)
	firstIP = cidr.Inc(firstIP)
	lastIP = cidr.Dec(lastIP)

	return firstIP, lastIP, nil
}

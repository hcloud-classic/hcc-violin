package iputil

import (
	"errors"
	"net"
	"strconv"
	"strings"
)

// CheckValidIP : Parses string value of IPv4 address then return as net.IP.
// If given wrong IP address, it wil return nil.
func CheckValidIP(ip string) net.IP {
	netIP := net.ParseIP(ip).To4()
	return netIP
}

// CheckNetmask : Check string value of IPv4 netmask then return as net.IPMask.
// If given wrong netmask, it will return nil and error.
func CheckNetmask(netmask string) (net.IPMask, error) {
	var err error

	var maskPartsStr = strings.Split(netmask, ".")
	if len(maskPartsStr) != 4 {
		return nil, errors.New("netmask should be X.X.X.X form")
	}

	var maskParts [4]int
	for i := range maskPartsStr {
		maskParts[i], err = strconv.Atoi(maskPartsStr[i])
		if err != nil {
			return nil, errors.New("netmask contained non-integer value")
		}
	}

	var mask = net.IPv4Mask(
		byte(maskParts[0]),
		byte(maskParts[1]),
		byte(maskParts[2]),
		byte(maskParts[3]))

	maskSizeOne, maskSizeBit := mask.Size()
	if maskSizeOne == 0 && maskSizeBit == 0 {
		return nil, errors.New("invalid netmask")
	}

	if maskSizeOne > 30 {
		return nil, errors.New("netmask bit should be equal or smaller than 30")
	}

	return mask, err
}

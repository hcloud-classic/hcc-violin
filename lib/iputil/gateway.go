package iputil

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
)

// Referenced from https://github.com/jackpal/gateway

// parseLinuxProcNetRoute parses the route file located at /proc/net/route
// and returns the IP address of the default gateway. The default gateway
// is the one with Destination value of 0.0.0.0.
//
// The Linux route file has the following format:
//
// $ cat /proc/net/route
//
// Iface   Destination Gateway     Flags   RefCnt  Use Metric  Mask
// eno1    00000000    C900A8C0    0003    0   0   100 00000000    0   00
// eno1    0000A8C0    00000000    0001    0   0   100 00FFFFFF    0   00
func parseLinuxProcNetRoute(f []byte) (net.IP, error) {
	const (
		sep              = "\t" // field separator
		destinationField = 1    // field containing hex destination address
		gatewayField     = 2    // field containing hex gateway address
	)
	scanner := bufio.NewScanner(bytes.NewReader(f))

	// Skip header line
	if !scanner.Scan() {
		return nil, errors.New("invalid linux route file")
	}

	var i = 0

	// make net.IP address from uint32
	ipd32 := make(net.IP, 4)

	for scanner.Scan() {
		row := scanner.Text()
		tokens := strings.Split(row, sep)
		if len(tokens) <= gatewayField {
			return nil, fmt.Errorf("invalid row '%s' in route file", row)
		}

		// Cast hex destination address to int
		destinationHex := "0x" + tokens[destinationField]
		destination, err := strconv.ParseInt(destinationHex, 0, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"parsing destination field hex '%s' in row '%s': %w",
				destinationHex,
				row,
				err,
			)
		}

		// The default interface is the one that's 0
		if destination == 0 {
			i++

			if i > 1 {
				return nil, errors.New("found multiple routes")
			}

			gatewayHex := "0x" + tokens[gatewayField]

			// cast hex address to uint32
			d, err := strconv.ParseInt(gatewayHex, 0, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"parsing default interface address field hex '%s' in row '%s': %w",
					destinationHex,
					row,
					err,
				)
			}
			d32 := uint32(d)

			binary.LittleEndian.PutUint32(ipd32, d32)
		}
	}

	if i == 0 {
		return nil, errors.New("interface with default destination not found")
	}

	// format net.IP to dotted ipV4 string
	return ipd32, nil
}

// GetDefaultRoute : Return IP of default route
func GetDefaultRoute() (net.IP, error) {
	var file = "/proc/net/route"

	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("can't access %s", file)
	}
	defer func() {
		_ = f.Close()
	}()

	readAll, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("can't read %s", file)
	}

	return parseLinuxProcNetRoute(readAll)
}

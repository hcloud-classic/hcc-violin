package harpUtil

import (
	"strconv"
	"strings"
)

// GetHarpVNUM : Calculate VNUM of the harp's interface
func GetHarpVNUM(networkAddress string) (vnum int) {
	var ifaceVNUM = 0

	ipSplit := strings.Split(networkAddress, ".")
	for _, ipSplited := range ipSplit {
		ipSplitedInt, _ := strconv.Atoi(ipSplited)
		ifaceVNUM += ipSplitedInt
	}

	return ifaceVNUM
}

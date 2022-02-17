package iputil

import (
	"testing"
)

func Test_CheckIP(t *testing.T) {
	netIP := CheckValidIP("192.168.100.0")
	if netIP == nil {
		t.Fatal("wrong network IP")
	}

	_, err := CheckNetmask("255.255.255.0")
	if err != nil {
		t.Fatal(err)
	}
}

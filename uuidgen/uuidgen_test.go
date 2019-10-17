package uuidgen

import (
	"hcc/violin/checkroot"
	"hcc/violin/config"
	"hcc/violin/logger"
	"hcc/violin/mysql"
	"testing"
)

func Test_UUIDgen(t *testing.T) {
	if !checkroot.CheckRoot() {
		t.Fatal("Failed to get root permission!")
	}

	if !logger.Prepare() {
		t.Fatal("Failed to prepare logger!")
	}
	defer func() {
		_ = logger.FpLog.Close()
	}()

	config.Parser()

	err := mysql.Prepare()
	if err != nil {
		return
	}
	defer func() {
		_ = mysql.Db.Close()
	}()

	_, err = UUIDgen()
	if err != nil {
		t.Fatal("Failed to generate uuid!")
	}
}

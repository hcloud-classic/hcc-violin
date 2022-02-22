package mysql

import (
	"hcc/violin/action/grpc/client"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"innogrid.com/hcloud-classic/hcc_errors"
	"testing"
)

func Test_DB_Prepare(t *testing.T) {
	err := logger.Init()
	if err != nil {
		hcc_errors.SetErrLogger(logger.Logger)
		hcc_errors.NewHccError(hcc_errors.ViolinInternalInitFail, "logger.Init(): "+err.Error()).Fatal()
	}
	hcc_errors.SetErrLogger(logger.Logger)
	if err != nil {
		t.Fatal("Failed to prepare logger!")
	}
	defer func() {
		_ = logger.FpLog.Close()
	}()

	config.Init()

	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}

	err = Init()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = Db.Close()
	}()
}

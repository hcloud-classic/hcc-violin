package mysql

import (
	"hcc/violin/lib/config"
	"hcc/violin/lib/errors"
	"hcc/violin/lib/logger"
	"testing"
)

func Test_DB_Prepare(t *testing.T) {
	err := logger.Init()
	if err != nil {
		errors.SetErrLogger(logger.Logger)
		errors.NewHccError(errors.ViolinInternalInitFail, "logger.Init(): "+err.Error()).Fatal()
	}
	errors.SetErrLogger(logger.Logger)
	if err != nil {
		t.Fatal("Failed to prepare logger!")
	}
	defer func() {
		_ = logger.FpLog.Close()
	}()

	config.Init()

	err = Init()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = Db.Close()
	}()
}

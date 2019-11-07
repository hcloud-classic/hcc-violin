package init

import (
	"errors"
	"hcc/violin/lib/logger"
)

func loggerInit() error {
	if !logger.Prepare() {
		return errors.New("error occurred while preparing logger")
	}
	defer func() {
		_ = logger.FpLog.Close()
	}()

	return nil
}

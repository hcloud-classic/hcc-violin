package end

import "hcc/violin/lib/logger"

func loggerEnd() {
	_ = logger.FpLog.Close()
}

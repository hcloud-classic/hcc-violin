package end

import "hcc/violin/lib/mysql"

func mysqlEnd() {
	if mysql.Db != nil {
		_ = mysql.Db.Close()
	}
}

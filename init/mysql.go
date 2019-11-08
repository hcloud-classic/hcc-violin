package init

import "hcc/violin/lib/mysql"

func mysqlInit() error {
	err := mysql.Prepare()
	if err != nil {
		return err
	}
	defer func() {
		if mysql.Db != nil {
			_ = mysql.Db.Close()
		}
	}()

	return nil
}

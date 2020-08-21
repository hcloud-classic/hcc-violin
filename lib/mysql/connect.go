package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql" // Needed for connect mysql
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
)

// Db : Pointer of mysql connection
var Db *sql.DB

// Init : Initialize mysql connection
func Init() error {
	var err error
	Db, err = sql.Open("mysql",
		config.Mysql.ID+":"+config.Mysql.Password+"@tcp("+
			config.Mysql.Address+":"+strconv.Itoa(int(config.Mysql.Port))+")/"+
			config.Mysql.Database+"?parseTime=true")
	if err != nil {
		logger.Logger.Println(err)
		return err
	}

	err = Db.Ping()
	if err != nil {
		logger.Logger.Println(err)
		return err
	}

	logger.Logger.Println("db is connected")

	return nil
}

// End : Close mysql connection
func End() {
	if Db != nil {
		_ = Db.Close()
	}
}

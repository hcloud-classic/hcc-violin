package uuidgen

import (
	"github.com/nu7hatch/gouuid"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
)

func checkDuplicateUUID(isServerNode bool, uuid string) (bool, error) {
	var sql string

	if isServerNode {
		sql = "select uuid from server"
	} else {
		sql = "select uuid from server_node"
	}
	stmt, err := mysql.Db.Query(sql)
	if err != nil {
		logger.Logger.Println(err)
		return false, err
	}
	defer func() {
		_ = stmt.Close()
	}()

	var dbUUID string
	var found = false
	for stmt.Next() {
		err := stmt.Scan(&dbUUID)
		if err != nil {
			logger.Logger.Println(err)
			return false, err
		}

		// logger.Logger.Printf("checkDuplicateUUID(): checking for UUID=%s\n", dbUUID)

		if uuid == dbUUID {
			logger.Logger.Println("checkDuplicateUUID(): Found already existed UUID")
			found = true
			break
		}
	}

	return found, nil
}

// UUIDgen : Generate uuid
func UUIDgen(isServerNode bool) (string, error) {
	var UUID string
	for {
		out, err := uuid.NewV4()
		if err != nil {
			logger.Logger.Println(err)
			return "", err
		}

		logger.Logger.Println("UUIDgen(): Checking duplicated UUID")

		var found = false
		if is_server_node {
			found, err = checkDuplicateUUID(true, out.String())
		} else {
			found, err = checkDuplicateUUID(false, out.String())
		}
		if err != nil {
			logger.Logger.Println(err)
			return "", err
		}

		if !found {
			UUID = out.String()
			break
		}
	}

	return UUID, nil
}

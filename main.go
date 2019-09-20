package main

import (
	"hcc/violin/config"
	"hcc/violin/graphql"
	"hcc/violin/logger"
	"hcc/violin/mysql"
	"net/http"
)

func main() {
	// if !checkroot.CheckRoot() {
	// 	return
	// }

	if !logger.Prepare() {
		return
	}
	defer logger.FpLog.Close()

	err := mysql.Prepare()
	if err != nil {
		return
	}
	defer mysql.Db.Close()

	http.Handle("/graphql", graphql.GraphqlHandler)

	logger.Log.Println("Server is running on port " + config.HTTPPort)
	err = http.ListenAndServe(":"+config.HTTPPort, nil)
	if err != nil {
		logger.Log.Println("Failed to prepare http server!")
	}
}

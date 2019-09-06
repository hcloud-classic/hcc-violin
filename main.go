package main

import (
	"GraphQL_violin/violincheckroot"
	"GraphQL_violin/violinconfig"
	"GraphQL_violin/violingraphql"
	"GraphQL_violin/violinlogger"
	"GraphQL_violin/violinmysql"
	"net/http"
)

func main() {
	if !violincheckroot.CheckRoot() {
		return
	}

	if !violinlogger.Prepare() {
		return
	}
	defer violinlogger.FpLog.Close()

	err := violinmysql.Prepare()
	if err != nil {
		return
	}
	defer violinmysql.Db.Close()

	http.Handle("/graphql", violingraphql.GraphqlHandler)

	violinlogger.Logger.Println("Server is running on port " + violinconfig.HTTPPort)
	err = http.ListenAndServe(":"+violinconfig.HTTPPort, nil)
	if err != nil {
		violinlogger.Logger.Println("Failed to prepare http server!")
	}
}

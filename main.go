package main

import (
	"hcc/violin/action/graphql"
	violinEnd "hcc/violin/end"
	violinInit "hcc/violin/init"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"net/http"
	"strconv"
)

func init() {
	err := violinInit.MainInit()
	if err != nil {
		panic(err)
	}
}

func main() {
	defer func() {
		violinEnd.MainEnd()
	}()

	http.Handle("/graphql", graphql.GraphqlHandler)
	err := http.ListenAndServe(":"+strconv.Itoa(int(config.HTTP.Port)), nil)
	if err != nil {
		logger.Logger.Println(err)
		logger.Logger.Println("Failed to prepare http server!")
		return
	}
	logger.Logger.Println("Server is running on port " + strconv.Itoa(int(config.HTTP.Port)))
}

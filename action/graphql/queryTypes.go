package graphql

import (
	"github.com/graphql-go/graphql"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
)

var queryTypes = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"server": &graphql.Field{
				Type:        serverType,
				Description: "Get server by uuid",
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: server")
					return dao.ReadServer(params.Args)
				},
			},
			"list_server": &graphql.Field{
				Type:        graphql.NewList(serverType),
				Description: "Get server list",
				Args: graphql.FieldConfigArgument{
					"subnet_uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"os": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"server_name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"server_desc": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"cpu": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"memory": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"disk_size": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"status": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"user_uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"row": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: list_server")
					return dao.ReadServerList(params.Args)
				},
			},
			"all_server": &graphql.Field{
				Type:        graphql.NewList(serverType),
				Description: "Get all server list",
				Args: graphql.FieldConfigArgument{
					"row": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: all_server")
					return dao.ReadServerAll(params.Args)
				},
			},
			"num_server": &graphql.Field{
				Type:        serverNum,
				Description: "Get the number of server",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: num_server")
					var serverNum model.ServerNum
					var err error
					serverNum, err = dao.ReadServerNum()

					return serverNum, err
				},
			},
		},
	})

package graphql

import (
	"github.com/graphql-go/graphql"
	graphqlType "hcc/violin/action/graphql/type"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
)

var queryTypes = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			// server DB
			"server": &graphql.Field{
				Type:        graphqlType.ServerType,
				Description: "Get server by uuid",
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"log_quiet": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					args := params.Args
					logQuiet, logQuietOk := args["log_quiet"].(bool)
					if !logQuietOk || !logQuiet {
						logger.Logger.Println("Resolving: server")
					}

					return dao.ReadServer(params.Args)
				},
			},
			"list_server": &graphql.Field{
				Type:        graphql.NewList(graphqlType.ServerType),
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
					logger.Logger.Println("Resolving: list_server")
					return dao.ReadServerList(params.Args)
				},
			},
			"all_server": &graphql.Field{
				Type:        graphql.NewList(graphqlType.ServerType),
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
					logger.Logger.Println("Resolving: all_server")
					return dao.ReadServerAll(params.Args)
				},
			},
			"num_server": &graphql.Field{
				Type:        graphqlType.ServerNumType,
				Description: "Get the number of server",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: num_server")
					var serverNum model.ServerNum
					var err error
					serverNum, err = dao.ReadServerNum()

					return serverNum, err
				},
			},
			// server_node DB
			"server_node": &graphql.Field{
				Type:        graphqlType.ServerNodeType,
				Description: "Get server_node by uuid",
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: server_node")
					return dao.ReadServerNode(params.Args)
				},
			},
			"list_server_node": &graphql.Field{
				Type:        graphql.NewList(graphqlType.ServerNodeType),
				Description: "Get server_node list",
				Args: graphql.FieldConfigArgument{
					"server_uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: list_server_node")
					return dao.ReadServerNodeList(params.Args)
				},
			},
			"all_server_node": &graphql.Field{
				Type:        graphql.NewList(graphqlType.ServerNodeType),
				Description: "Get all server_node list",
				Args:        graphql.FieldConfigArgument{},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: all_server_node")
					return dao.ReadServerNodeAll()
				},
			},
			"num_nodes_server": &graphql.Field{
				Type:        graphqlType.ServerNumType,
				Description: "Get the number of nodes of server",
				Args: graphql.FieldConfigArgument{
					"server_uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: num_nodes_server")
					return dao.ReadServerNodeNum(params.Args)
				},
			},
		},
	})

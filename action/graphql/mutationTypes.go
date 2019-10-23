package graphql

import (
	"github.com/graphql-go/graphql"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/uuidgen"
)

var mutationTypes = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		// server DB
		"create_server": &graphql.Field{
			Type:        graphql.String,
			Description: "Create new server",
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
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				serverUUID, err := uuidgen.UUIDgen()
				if err != nil {
					logger.Logger.Println("Failed to generate uuid!")
					return "", err
				}

				// stage 1. select node - reader, compute
				listNodeData, err := GetNodes()
				if err != nil {
					logger.Logger.Print(err)
					return "", err
				}

				// stage 1.1 update nodes info (server_uuid)
				var nodes = listNodeData.Data.ListNode
				for _, node := range nodes {
					err = UpdateNode(node, serverUUID)
					if err != nil {
						logger.Logger.Println(err)
					}
				}

				// stage

				// stage 2. create volume - os, data

				// stage 3. create subnet

				// stage 4. node power on

				// stage 5. viola install

				return dao.CreateServer(serverUUID, params.Args)
			},
		},
		"update_server": &graphql.Field{
			Type:        serverType,
			Description: "Update server",
			Args: graphql.FieldConfigArgument{
				"uuid": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
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
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				logger.Logger.Println("Resolving: update_server")
				return dao.UpdateServer(params.Args)
			},
		},
		"delete_server": &graphql.Field{
			Type:        serverType,
			Description: "Delete server by uuid",
			Args: graphql.FieldConfigArgument{
				"uuid": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				logger.Logger.Println("Resolving: delete_volume")
				return dao.DeleteServer(params.Args)
			},
		},
		// server_node DB
		"create_server_node": &graphql.Field{
			Type:        serverNodeType,
			Description: "Create new server_node",
			Args: graphql.FieldConfigArgument{
				"server_uuid": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"node_uuid": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				return dao.CreateServerNode(params.Args)
			},
		},
		"delete_server_node": &graphql.Field{
			Type:        serverNodeType,
			Description: "Delete server_node by uuid",
			Args: graphql.FieldConfigArgument{
				"uuid": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				logger.Logger.Println("Resolving: delete server_node")
				return dao.DeleteServerNode(params.Args)
			},
		},
	},
})

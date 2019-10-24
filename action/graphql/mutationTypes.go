package graphql

import (
	"github.com/graphql-go/graphql"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/uuidgen"
	"hcc/violin/model"
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
				serverUUID, err := uuidgen.UUIDgen(false)
				if err != nil {
					logger.Logger.Println("Failed to generate uuid!")
					return "", err
				}

				userUUID := params.Args["user_uuid"].(string)
				diskSize := params.Args["disk_size"].(int)

				// stage 1. select node - reader, compute
				listNodeData, err := GetNodes()
				if err != nil {
					logger.Logger.Print(err)
					return "", err
				}

				// stage 1.1 update nodes info (server_uuid)
				// stage 1.2 insert nodes to server_node table
				var nodes = listNodeData.Data.ListNode
				for _, node := range nodes {
					err = UpdateNode(node, serverUUID)
					if err != nil {
						logger.Logger.Println(err)
						return "", err
					}

					args := make(map[string]interface{})
					args["server_uuid"] = serverUUID
					args["node_uuid"] = node.UUID
					_, err = dao.CreateServerNode(args)
					if err != nil {
						logger.Logger.Println(err)
						return "", err
					}
				}

				// stage 2. create volume - os, data
				var volumeOS = model.Volume{
					Size:       model.OSDiskSize,
					Filesystem: model.DefaultPXEdir + "/" + serverUUID,
					ServerUUID: serverUUID,
					UseType:    "os",
					UserUUID:   userUUID,
				}
				err = CreateDisk(volumeOS, serverUUID)
				if err != nil {
					logger.Logger.Println(err)
					return "", err
				}

				var volumeData = model.Volume{
					Size:       diskSize,
					Filesystem: model.DefaultPXEdir + "/" + serverUUID,
					ServerUUID: serverUUID,
					UseType:    "data",
					UserUUID:   userUUID,
				}
				err = CreateDisk(volumeData, serverUUID)
				if err != nil {
					logger.Logger.Println(err)
					return "", err
				}

				// stage 3. create dhcpd conf (get subnet info -> create dhcpd config)
				// stage 3.1 restart dhcpd service

				// stage 4. node power on

				// stage 5. viola install
				// RunHccCLI(xxx)
				// while checking Cello DB cluster status is runnig in N times, until retry is expired

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

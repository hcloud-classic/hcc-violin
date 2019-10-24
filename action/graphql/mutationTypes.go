package graphql

import (
	"hcc/violin/action/rabbitmq"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/uuidgen"
	"hcc/violin/model"
	"time"

	"github.com/graphql-go/graphql"
)

var mutationTypes = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		// server DB
		"create_server": &graphql.Field{
			Type:        serverType,
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
					return nil, err
				}

				userUUID := params.Args["user_uuid"].(string)
				os := params.Args["os"].(string)
				diskSize := params.Args["disk_size"].(int)

				// stage 1. select node - reader, compute
				listNodeData, err := GetNodes()
				if err != nil {
					logger.Logger.Print(err)
					return nil, err
				}

				// stage 1.1 update nodes info (server_uuid)
				// stage 1.2 insert nodes to server_node table
				var nodes = listNodeData.Data.ListNode
				for _, node := range nodes {
					err = UpdateNode(node, serverUUID)
					if err != nil {
						logger.Logger.Println(err)
						return nil, err
					}

					args := make(map[string]interface{})
					args["server_uuid"] = serverUUID
					args["node_uuid"] = node.UUID
					_, err = dao.CreateServerNode(args)
					if err != nil {
						logger.Logger.Println(err)
						return nil, err
					}
				}

				go func() {
					subnetUUID := params.Args["subnet_uuid"].(string)
					subnet, err := GetSubnet(subnetUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}

					// stage 2. create volume - os, data
					var volumeOS = model.Volume{
						Size:       model.OSDiskSize,
						Filesystem: os,
						ServerUUID: serverUUID,
						UseType:    "os",
						UserUUID:   userUUID,
						NetworkIP:  subnet.Data.Subnet.NetworkIP,
					}
					err = CreateDisk(volumeOS, serverUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}

					var volumeData = model.Volume{
						Size:       diskSize,
						Filesystem: os,
						ServerUUID: serverUUID,
						UseType:    "data",
						UserUUID:   userUUID,
						NetworkIP:  subnet.Data.Subnet.NetworkIP,
					}
					err = CreateDisk(volumeData, serverUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}

					// stage 3. UpdateSubnet (get subnet info -> create dhcpd config -> update_subnet)
					_, err = UpdateSubnet(subnetUUID, serverUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}

					// stage 4. node power on
					for _, node := range nodes {
						if subnet.Data.Subnet.LeaderNodeUUID == node.UUID {
							result, err := OnNode(node.PXEMacAddr)
							if err != nil {
								logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
								return
							}

							logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": , OnNode leader MAC Addr: " + node.PXEMacAddr + result)
						}
					}

					// Wail for leader node to turn on for 100secs
					time.Sleep(40 * time.Second)

					for _, node := range nodes {
						if subnet.Data.Subnet.LeaderNodeUUID == node.UUID {
							continue
						}

						result, err := OnNode(node.PXEMacAddr)
						if err != nil {
							logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
							return
						}

						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": , OnNode leader MAC Addr: " + node.PXEMacAddr + result)
					}
					var controlAction = model.Control{

						HccCommand: "hcc nodes add -n 0",
						HccIPRange: subnet.Data.Subnet.NetworkIP,
						ServerUUID: serverUUID,
					}
					// stage 5. viola install
					rabbitmq.RunHccCLI(controlAction)
					// while checking Cello DB cluster status is runnig in N times, until retry is expired

				}()

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

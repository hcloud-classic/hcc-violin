package graphql

import (
	"hcloud-violin/logger"
	"hcloud-violin/mysql"
	"hcloud-violin/types"
	"time"

	"github.com/graphql-go/graphql"
)

var queryTypes = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			////////////////////////////// server ///////////////////////////////
			/* Get (read) single server by uuid
			   http://localhost:7500/graphql?query={server(uuid:"[server_uuid]]"){uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid, createdAt}}
			*/
			"server": &graphql.Field{
				Type:        serverType,
				Description: "Get server by uuid",
				Args: graphql.FieldConfigArgument{
					"uuid": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: server")

					requestedUUID, ok := p.Args["uuid"].(string)
					if ok {
						server := new(types.Server)

						var uuid string
						var subnetUUID string
						var os string
						var serverName string
						var serverDesc string
						var cpu int
						var memory int
						var diskSize int
						var status string
						var userUUID string
						var createdAt time.Time

						sql := "select * from server where uuid = ?"
						err := mysql.Db.QueryRow(sql, requestedUUID).Scan(&uuid, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
						if err != nil {
							logger.Log.Println(err)
							return nil, nil
						}

						server.UUID = uuid
						server.SubnetUUID = subnetUUID
						server.OS = os
						server.ServerName = serverName
						server.ServerDesc = serverDesc
						server.CPU = cpu
						server.Memory = memory
						server.DiskSize = diskSize
						server.Status = status
						server.UserUUID = userUUID
						server.CreatedAt = createdAt

						return server, nil
					}
					return nil, nil
				},
			},

			/* Get (read) server list
			   http://localhost:7500/graphql?query={list_server{uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid}}
			*/
			"list_server": &graphql.Field{
				Type:        graphql.NewList(serverType),
				Description: "Get server list",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: list_server")

					var servers []types.Server
					var uuid string
					var subnetUUID string
					var os string
					var serverName string
					var serverDesc string
					var cpu int
					var memory int
					var diskSize int
					var status string
					var userUUID string
					var createdAt time.Time

					sql := "select * from server"
					stmt, err := mysql.Db.Query(sql)
					if err != nil {
						logger.Log.Println(err)
						return nil, nil
					}
					defer stmt.Close()

					for stmt.Next() {
						err := stmt.Scan(&uuid, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
						if err != nil {
							logger.Log.Println(err)
						}

						server := types.Server{UUID: uuid, SubnetUUID: subnetUUID, OS: os, ServerName: serverName, ServerDesc: serverDesc, CPU: cpu, Memory: memory, DiskSize: diskSize, Status: status, UserUUID: userUUID, CreatedAt: createdAt}

						logger.Log.Println(server)
						servers = append(servers, server)
					}

					return servers, nil
				},
			},
		},
	})

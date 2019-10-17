package graphql

import (
	"github.com/graphql-go/graphql"
	"hcc/violin/logger"
	"hcc/violin/mysql"
	"hcc/violin/types"
	"time"
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
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: server")

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
							logger.Logger.Println(err)
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
			"list_server": &graphql.Field{
				Type:        graphql.NewList(serverType),
				Description: "Get server list",
				Args: graphql.FieldConfigArgument{
					"row": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"page": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},

				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: list_server")

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
					var row int
					var page int
					row = params.Args["row"].(int)
					page = params.Args["page"].(int)

					sql := "select * from server order by created_at desc limit ? offset ?"
					stmt, err := mysql.Db.Query(sql, row, row*(page-1))
					if err != nil {
						logger.Logger.Println(err.Error())
						return nil, nil
					}
					defer func() {
						_ = stmt.Close()
					}()

					for stmt.Next() {
						err := stmt.Scan(&uuid, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
						if err != nil {
							logger.Logger.Println(err)
						}
						server := types.Server{UUID: uuid, SubnetUUID: subnetUUID, OS: os, ServerName: serverName, ServerDesc: serverDesc, CPU: cpu, Memory: memory, DiskSize: diskSize, Status: status, UserUUID: userUUID, CreatedAt: createdAt}
						logger.Logger.Println(server)
						servers = append(servers, server)
					}
					return servers, nil
				},
			},
			"num_server": &graphql.Field{
				Type:        serverNum,
				Description: "Get the number of server",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Logger.Println("Resolving: list_server")

					var serverNum types.ServerNum
					var serverNr int

					sql := "select count(*) from server"
					err := mysql.Db.QueryRow(sql).Scan(&serverNr)
					if err != nil {
						logger.Logger.Println(err)
						return nil, nil
					}
					logger.Logger.Println("Count: ", serverNr)
					serverNum.Number = serverNr

					return serverNum, nil
				},
			},
		},
	})

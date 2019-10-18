package graphql

import (
	"github.com/graphql-go/graphql"
	"hcc/violin/logger"
	"hcc/violin/mysql"
	"hcc/violin/types"
	"strconv"
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

					var servers []types.Server
					var rxUUID string
					var createdAt time.Time

					subnetUUID, subnetUUIDOk := params.Args["subnet_uuid"].(string)
					os, osOk := params.Args["os"].(string)
					serverName, serverNameOk := params.Args["server_name"].(string)
					serverDesc, serverDescOk := params.Args["server_desc"].(string)
					cpu, cpuOk := params.Args["cpu"].(int)
					memory, memoryOk := params.Args["memory"].(int)
					diskSize, diskSizeOk := params.Args["disk_size"].(int)
					status, statusOk := params.Args["status"].(string)
					userUUID, userUUIDOk := params.Args["user_uuid"].(string)
					if !userUUIDOk {
						return nil, nil
					}
					row, rowOk := params.Args["row"].(int)
					page, pageOk := params.Args["page"].(int)

					if !rowOk || !pageOk {
						return nil, nil
					}

					sql := "select * from server where"
					if subnetUUIDOk {
						sql += " subnet_uuid = '" + subnetUUID + "'"
						if osOk || serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if osOk {
						sql += " os = '" + os + "'"
						if serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if serverNameOk {
						sql += " server_name = '" + serverName + "'"
						if serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if serverDescOk {
						sql += " server_desc = '" + serverDesc + "'"
						if cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if cpuOk {
						sql += " cpu = " + strconv.Itoa(cpu)
						if memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if memoryOk {
						sql += " memory = " + strconv.Itoa(memory)
						if diskSizeOk || statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if diskSizeOk {
						sql += " disk_size = " + strconv.Itoa(diskSize)
						if statusOk || userUUIDOk {
							sql += " and"
						}
					}
					if statusOk {
						sql += " status = '" + status + "' and"
					}

					sql += " user_uuid = ? order by created_at desc limit ? offset ?"

					logger.Log.Println("list_server sql  : ", sql)

					stmt, err := mysql.Db.Query(sql, userUUID, row, row*(page-1))
					if err != nil {
						logger.Log.Println(err.Error())
						return nil, nil
					}
					defer stmt.Close()

					for stmt.Next() {
						err := stmt.Scan(&rxUUID, &subnetUUID, &os, &serverName, &serverDesc, &cpu, &memory, &diskSize, &status, &userUUID, &createdAt)
						if err != nil {
							logger.Log.Println(err)
						}
						server := types.Server{UUID: rxUUID, SubnetUUID: subnetUUID, OS: os, ServerName: serverName, ServerDesc: serverDesc, CPU: cpu, Memory: memory, DiskSize: diskSize, Status: status, UserUUID: userUUID, CreatedAt: createdAt}
						logger.Log.Println(server)
						servers = append(servers, server)
					}
					return servers, nil
				},
			},
			"all_server": &graphql.Field{
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
					row, rowOk := params.Args["row"].(int)
					page, pageOk := params.Args["page"].(int)

					if !rowOk || !pageOk {
						return nil, nil
					}

					sql := "select * from server order by created_at desc limit ? offset ?"
					logger.Log.Println("list_server sql  : ", sql)

					stmt, err := mysql.Db.Query(sql, row, row*(page-1))
					if err != nil {
						logger.Log.Println(err.Error())
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
			"num_server": &graphql.Field{
				Type:        serverNum,
				Description: "Get the number of server",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					logger.Log.Println("Resolving: list_server")

					var serverNum types.ServerNum
					var serverNr int

					sql := "select count(*) from server"
					err := mysql.Db.QueryRow(sql).Scan(&serverNr)
					if err != nil {
						logger.Log.Println(err)
						return nil, nil
					}
					logger.Log.Println("Count: ", serverNr)
					serverNum.Number = serverNr

					return serverNum, nil
				},
			},
		},
	})

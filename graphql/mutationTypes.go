package graphql

import (
	"github.com/graphql-go/graphql"
	"hcc/violin/logger"
	"hcc/violin/mysql"
	"hcc/violin/types"
	"hcc/violin/uuidgen"
	"strconv"
)

var mutationTypes = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
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
				logger.Log.Println("Resolving: create_server")
				uuid, err := uuidgen.Uuidgen()
				if err != nil {
					logger.Log.Println("Failed to generate uuid!")
					return nil, nil
				}

				server := types.Server{
					UUID:       uuid,
					SubnetUUID: params.Args["subnet_uuid"].(string),
					OS:         params.Args["os"].(string),
					ServerName: params.Args["server_name"].(string),
					ServerDesc: params.Args["server_desc"].(string),
					CPU:        params.Args["cpu"].(int),
					Memory:     params.Args["memory"].(int),
					DiskSize:   params.Args["disk_size"].(int),
					Status:     params.Args["status"].(string),
					UserUUID:   params.Args["user_uuid"].(string),
				}

				sql := "insert into server(uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid, created_at) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())"
				stmt, err := mysql.Db.Prepare(sql)
				if err != nil {
					logger.Log.Println(err.Error())
					return nil, nil
				}
				defer stmt.Close()
				result, err := stmt.Exec(server.UUID, server.SubnetUUID, server.OS, server.ServerName, server.ServerDesc, server.CPU, server.Memory, server.DiskSize, server.Status, server.UserUUID)
				if err != nil {
					logger.Log.Println(err)
					return nil, nil
				}
				logger.Log.Println(result.LastInsertId())

				// stage 1. select node - reader, compute

				// stage 2. create volume - os, data

				// stage 3. create subnet

				// stage 4. node power on

				// stage 5. viola install

				return server, nil
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
				logger.Log.Println("Resolving: update_server")

				requestedUUID, requestedUUIDOk := params.Args["uuid"].(string)
				subnetUUID, subnetUUIDOk := params.Args["subnet_uuid"].(string)
				os, osOk := params.Args["os"].(string)
				serverName, serverNameOk := params.Args["server_name"].(string)
				serverDesc, serverDescOk := params.Args["server_desc"].(string)
				cpu, cpuOk := params.Args["cpu"].(int)
				memory, memoryOk := params.Args["memory"].(int)
				diskSize, diskSizeOk := params.Args["disk_size"].(int)
				status, statusOk := params.Args["status"].(string)
				userUUID, userUUIDOk := params.Args["user_uuid"].(string)

				server := new(types.Server)
				server.UUID = requestedUUID
				server.SubnetUUID = subnetUUID
				server.OS = os
				server.ServerName = serverName
				server.ServerDesc = serverDesc
				server.CPU = cpu
				server.Memory = memory
				server.DiskSize = diskSize
				server.Status = status
				server.UserUUID = userUUID

				if requestedUUIDOk {
					if !subnetUUIDOk && !osOk && !serverNameOk && !serverDescOk && !cpuOk && !memoryOk && !diskSizeOk && !statusOk && !userUUIDOk {
						return nil, nil
					}
					sql := "update server set"
					if subnetUUIDOk {
						sql += " subnet_uuid = '" + server.SubnetUUID + "'"
						if osOk || serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if osOk {
						sql += " os = '" + server.OS + "'"
						if serverNameOk || serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if serverNameOk {
						sql += " server_name = '" + server.ServerName + "'"
						if serverDescOk || cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if serverDescOk {
						sql += " server_desc = '" + server.ServerDesc + "'"
						if cpuOk || memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if cpuOk {
						sql += " cpu = " + strconv.Itoa(server.CPU)
						if memoryOk || diskSizeOk || statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if memoryOk {
						sql += " memory = " + strconv.Itoa(server.Memory)
						if diskSizeOk || statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if diskSizeOk {
						sql += " disk_size = " + strconv.Itoa(server.DiskSize)
						if statusOk || userUUIDOk {
							sql += ", "
						}
					}
					if statusOk {
						sql += " status = '" + server.Status + "'"
						if userUUIDOk {
							sql += ", "
						}
					}
					if userUUIDOk {
						sql += " user_uuid = " + server.UserUUID
					}
					sql += " where uuid = ?"

					logger.Log.Println("update_server sql : ", sql)

					stmt, err := mysql.Db.Prepare(sql)
					if err != nil {
						logger.Log.Println(err.Error())
						return nil, nil
					}
					defer stmt.Close()

					result, err2 := stmt.Exec(server.UUID)
					if err2 != nil {
						logger.Log.Println(err2)
						return nil, nil
					}
					logger.Log.Println(result.LastInsertId())
					return server, nil
				}
				return nil, nil
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
				logger.Log.Println("Resolving: delete_volume")

				requestedUUID, ok := params.Args["uuid"].(string)
				if ok {
					sql := "delete from server where uuid = ?"
					stmt, err := mysql.Db.Prepare(sql)
					if err != nil {
						logger.Log.Println(err.Error())
						return nil, nil
					}
					defer stmt.Close()
					result, err2 := stmt.Exec(requestedUUID)
					if err2 != nil {
						logger.Log.Println(err2)
						return nil, nil
					}
					logger.Log.Println(result.RowsAffected())

					return requestedUUID, nil
				}
				return nil, nil
			},
		},
	},
})

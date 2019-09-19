package graphql

import (
	"hcloud-violin/logger"
	"hcloud-violin/mysql"
	"hcloud-violin/types"
	"hcloud-violin/uuidgen"

	"github.com/graphql-go/graphql"
)

var mutationTypes = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		////////////////////////////// server ///////////////////////////////
		/* Create new server
		http://localhost:7500/graphql?query=mutation+_{create_server(size:1024000,type:"ext4",server_uuid:"[server_uuid]"){uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid}}
		*/
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

				sql := "insert into server(uuid, subnet_uuid, os, server_name, server_desc, cpu, memory, disk_size, status, user_uuid) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
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

				return server, nil
			},
		},

		/* Update server by uuid
		   http://localhost:8001/graphql?query=mutation+_{update_volume(uuid:"[volume_uuid]",size:10240,type:"ext4",server_uuid:"[server_uuid]"){uuid,size,type,server_uuid}}
		*/
		"update_server": &graphql.Field{
			Type:        serverType,
			Description: "Update server by uuid",
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

				requestedUUID, _ := params.Args["uuid"].(string)
				subnetUUID, subnetUUIDOK := params.Args["subnet_uuid"].(string)
				os, osOK := params.Args["os"].(string)
				serverName, serverNameOK := params.Args["server_name"].(string)
				serverDesc, serverDescOK := params.Args["server_desc"].(string)
				cpu, cpuOK := params.Args["cpu"].(int)
				memory, memoryOK := params.Args["memory"].(int)
				diskSize, diskSizeOK := params.Args["disk_size"].(int)
				status, statusOK := params.Args["status"].(string)
				userUUID, userUUIDOK := params.Args["user_uuid"].(string)

				server := new(types.Server)

				if subnetUUIDOK && osOK && serverNameOK && serverDescOK && cpuOK && memoryOK && diskSizeOK && statusOK && userUUIDOK {
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

					sql := "update server set subnet_uuid = ?, os = ?, server_name = ?, server_desc= ?, cpu=?, memory=?, disk_size=?, status=?, user_uuid=?  where uuid = ?"
					stmt, err := mysql.Db.Prepare(sql)
					if err != nil {
						logger.Log.Println(err.Error())
						return nil, nil
					}
					defer stmt.Close()
					result, err2 := stmt.Exec(server.SubnetUUID, server.OS, server.ServerName, server.ServerDesc, server.CPU, server.Memory, server.DiskSize, server.Status, server.UserUUID, server.UUID)
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

		/* Delete server by id
		   http://localhost:8001/graphql?query=mutation+_{delete_volume(id:"test1"){id}}
		*/
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

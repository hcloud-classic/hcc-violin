package graphql

import (
	"errors"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/graphql-go/graphql"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/uuidgen"
	"hcc/violin/model"
	"net"
	"strconv"
	"strings"
	"time"
)

func checkNetmask(netmask string) (net.IPMask, error) {
	var err error

	var maskPartsStr = strings.Split(netmask, ".")
	if len(maskPartsStr) != 4 {
		return nil, errors.New("netmask should be X.X.X.X form")
	}

	var maskParts [4]int
	for i := range maskPartsStr {
		maskParts[i], err = strconv.Atoi(maskPartsStr[i])
		if err != nil {
			return nil, errors.New("netmask contained none integer value")
		}
	}

	var mask = net.IPv4Mask(
		byte(maskParts[0]),
		byte(maskParts[1]),
		byte(maskParts[2]),
		byte(maskParts[3]))

	maskSizeOne, maskSizeBit := mask.Size()
	if maskSizeOne == 0 && maskSizeBit == 0 {
		return nil, errors.New("invalid netmask")
	}

	if maskSizeOne > 30 {
		return nil, errors.New("netmask bit should be equal or smaller than 30")
	}

	return mask, err
}

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

				// stage 1. select node - leader, compute
				listNodeData, err := GetNodes()
				if err != nil {
					logger.Logger.Print(err)
					return nil, err
				}

				// TODO : Currently nrNodes is hard coded to 2. Will get from Web UI (Oboe) later.
				var nrNodes = 2

				// stage 1.1 update nodes info (server_uuid)
				// stage 1.2 insert nodes to server_node table
				var nodes = listNodeData.Data.ListNode

				if len(nodes) < nrNodes {
					return nil, errors.New("not enough available nodes")
				}

				var nodeSelected = 0
				for _, node := range nodes {
					if nodeSelected > nrNodes {
						break
					}

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

					nodeSelected++
				}

				go func() {
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Getting subnet info")

					subnetUUID := params.Args["subnet_uuid"].(string)
					subnet, err := GetSubnet(subnetUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}

					// stage 2. create volume - os, data
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Creating os volume")
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

					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Creating data volume")
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
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Updating subnet info")
					_, err = UpdateSubnet(subnetUUID, serverUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}

					// stage 4. node power on
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Turning on leader node")
					_, err = OnNode(subnet.Data.Subnet.LeaderNodeUUID)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
						return
					}
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": , OnNode leader MAC Addr: " + "d0-50-99-aa-e5-7b")

					//fmt.Println("leader: " + subnet.Data.Subnet.LeaderNodeUUID)
					//for _, node := range nodes {
					//	fmt.Println("node: " + node.UUID)
					//	if subnet.Data.Subnet.LeaderNodeUUID == node.UUID {
					//		result, err := OnNode(node.PXEMacAddr)
					//		if err != nil {
					//			logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
					//			return
					//		}
					//
					//	}
					//}

					// Wait for leader node to turn on for 30secs
					time.Sleep(time.Second * 30)

					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Turning on compute nodes")
					for _, node := range nodes {
						if node.UUID == subnet.Data.Subnet.LeaderNodeUUID {
							continue
						}

						_, err := OnNode(node.PXEMacAddr)
						if err != nil {
							logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
							return
						}

						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": , OnNode compute MAC Addr: " + node.PXEMacAddr)
					}

					netIPnetworkIP := net.ParseIP(subnet.Data.Subnet.NetworkIP).To4()
					if netIPnetworkIP == nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "got wrong network IP")
						return
					}

					mask, err := checkNetmask(subnet.Data.Subnet.Netmask)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "got wrong subnet mask")
						return
					}

					ipNet := net.IPNet{
						IP:   netIPnetworkIP,
						Mask: mask,
					}

					firstIP, _ := cidr.AddressRange(&ipNet)
					firstIP = cidr.Inc(firstIP)
					lastIP := firstIP

					for i := 0; i < len(nodes)-1; i++ {
						lastIP = cidr.Inc(lastIP)
					}

					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Preparing controlAction")

					var controlAction = model.Control{
						HccCommand: "hcc nodes add -n 2",
						HccIPRange: "range " + firstIP.String() + " " + lastIP.String(),
						ServerUUID: serverUUID,
					}

					// stage 5. viola install
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Running HccCLI")

					err = rabbitmq.ViolinToViola(controlAction)
					if err != nil {
						logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
						return
					}
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

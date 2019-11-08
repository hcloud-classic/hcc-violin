package graphql

import (
	"errors"
	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/graphql-go/graphql"
	uuid "github.com/nu7hatch/gouuid"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/dao"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"net"
	"time"
)

func createServer(params graphql.ResolveParams) (interface{}, error) {
	logger.Logger.Println("createServer: Getting subnet info from harp module")

	subnetUUID := params.Args["subnet_uuid"].(string)
	subnet, err := GetSubnet(subnetUUID)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}

	if len(subnet.ServerUUID) != 0 {
		errMsg := "createServer: Selected subnet (subnetUUID=" + subnetUUID +
			") is used by one of server (serverUUID=" + subnet.ServerUUID + ")"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}
	logger.Logger.Println("createServer: subnet info: network IP=" + subnet.NetworkIP +
		", netmask=" + subnet.Netmask)

	netIPnetworkIP := net.ParseIP(subnet.NetworkIP).To4()
	if netIPnetworkIP == nil {
		errMsg := "createServer: got wrong network IP"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	mask, err := checkNetmask(subnet.Netmask)
	if err != nil {
		errMsg := "createServer: got wrong subnet mask"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	ipNet := net.IPNet{
		IP:   netIPnetworkIP,
		Mask: mask,
	}

	logger.Logger.Println("createServer: Generating server UUID")
	out, err := uuid.NewV4()
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}
	serverUUID := out.String()

	userUUID := params.Args["user_uuid"].(string)
	os := params.Args["os"].(string)
	diskSize := params.Args["disk_size"].(int)

	// stage 1. select node - leader, compute
	logger.Logger.Println("createServer: Getting available nodes from flute module")

	listNodeData, err := GetNodes()
	nodes := listNodeData.(ListNodeData).Data.ListNode
	if err != nil {
		logger.Logger.Print(err)
		return nil, err
	}

	// TODO : Currently nrNodes is hard coded to 2. Will get from Web UI (Oboe) later.
	var nrNodes = 2

	// TODO : Get leader node's UUID from selected nodes. Currently, leader node's UUID is provided by subnet data.
	// stage 1.1 update nodes info (server_uuid)
	// stage 1.2 insert nodes to server_node table
	var nodeUUIDs []string

	if len(nodes) < nrNodes || len(nodes) == 0 {
		errMsg := "createServer: not enough available nodes"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	var nodeSelected = 0
	for _, node := range nodes {
		if nodeSelected > nrNodes {
			break
		}

		logger.Logger.Println("createServer: Updating nodes info to flute module")

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

		nodeUUIDs = append(nodeUUIDs, node.UUID)

		nodeSelected++
	}

	logger.Logger.Println("createServer: Getting IP address range")
	firstIP, _ := cidr.AddressRange(&ipNet)
	firstIP = cidr.Inc(firstIP)
	lastIP := firstIP

	for i := 0; i < len(nodes)-1; i++ {
		lastIP = cidr.Inc(lastIP)
	}

	go func() {
		// stage 2. create volume - os, data
		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Creating os volume")
		var volumeOS = model.Volume{
			Size:       model.OSDiskSize,
			Filesystem: os,
			ServerUUID: serverUUID,
			UseType:    "os",
			UserUUID:   userUUID,
			NetworkIP:  subnet.NetworkIP,
		}
		err = CreateDisk(volumeOS, serverUUID)
		if err != nil {
			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + err.Error())
			return
		}

		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Creating data volume")
		var volumeData = model.Volume{
			Size:       diskSize,
			Filesystem: os,
			ServerUUID: serverUUID,
			UseType:    "data",
			UserUUID:   userUUID,
			NetworkIP:  subnet.NetworkIP,
		}
		err = CreateDisk(volumeData, serverUUID)
		if err != nil {
			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + err.Error())
			return
		}

		// stage 3. UpdateSubnet (get subnet info -> create dhcpd config -> update_subnet)
		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Updating subnet info")
		_, err = UpdateSubnet(subnetUUID, serverUUID)
		if err != nil {
			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + " UpdateSubnet: " + err.Error())
			return
		}

		var nodeUUIDsStr = ""
		for i, node := range nodes {
			nodeUUIDsStr += node.UUID
			if i != len(nodes)-1 {
				nodeUUIDsStr += ","
			}
		}
		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + " nodeUUIDsStr: " + nodeUUIDsStr)

		err = CreateDHCPDConfig(subnetUUID, nodeUUIDsStr)
		if err != nil {
			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + " CreateDHCPDConfig: " + err.Error())
			return
		}

		// stage 4. node power on
		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Turning on leader node")
		var i = 1
		for _, node := range nodes {
			if node.UUID == subnet.LeaderNodeUUID {
				_, err := OnNode(node.PXEMacAddr)
				if err != nil {
					logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
					return
				}

				logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": OnNode: leader MAC Addr: " + node.PXEMacAddr)

				break
			}

			i++
		}

		if i > len(nodes) {
			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Failed to find leader node")
			return
		}

		// Wait for leader node to turned on
		time.Sleep(time.Second * time.Duration(config.Flute.WaitForLeaderNodeTimeoutSec))

		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Turning on compute nodes")
		for _, node := range nodes {
			if node.UUID == subnet.LeaderNodeUUID {
				continue
			}

			_, err := OnNode(node.PXEMacAddr)
			if err != nil {
				logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
				return
			}

			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": OnNode: compute MAC Addr: " + node.PXEMacAddr)
		}

		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Preparing controlAction")

		var controlAction = model.Control{
			HccCommand: "hcc nodes add -n 2",
			HccIPRange: "range " + firstIP.String() + " " + lastIP.String(),
			ServerUUID: serverUUID,
		}

		// stage 5. viola install
		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + "Running HccCLI")

		err = rabbitmq.ViolinToViola(controlAction)
		if err != nil {
			logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": " + err.Error())
			return
		}
		// while checking Cello DB cluster status is runnig in N times, until retry is expired
	}()

	return dao.CreateServer(serverUUID, params.Args)
}

func updateServer(params graphql.ResolveParams) (interface{}, error) {
	// TODO : Update server stages

	return dao.UpdateServer(params.Args)
}

func deleteServer(params graphql.ResolveParams) (interface{}, error) {
	// TODO : Delete server stages

	return dao.DeleteServer(params.Args)
}

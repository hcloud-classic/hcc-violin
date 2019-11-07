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
	logger.Logger.Println("create_server: Getting subnet info from harp module")

	subnetUUID := params.Args["subnet_uuid"].(string)
	subnet, err := ToHarpGetSubnet(subnetUUID)
	if err != nil {
		logger.Logger.Println(err)
		return nil, err
	}

	if len(subnet.Data.Subnet.ServerUUID) != 0 {
		errMsg := "create_server: Selected subnet (subnetUUID=" + subnetUUID +
			") is used by one of server (serverUUID=" + subnet.Data.Subnet.ServerUUID + ")"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}
	logger.Logger.Println("create_server: subnet info: network IP=" + subnet.Data.Subnet.NetworkIP +
		", netmask=" + subnet.Data.Subnet.Netmask)

	netIPnetworkIP := net.ParseIP(subnet.Data.Subnet.NetworkIP).To4()
	if netIPnetworkIP == nil {
		errMsg := "create_server: got wrong network IP"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	mask, err := checkNetmask(subnet.Data.Subnet.Netmask)
	if err != nil {
		errMsg := "create_server: got wrong subnet mask"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	ipNet := net.IPNet{
		IP:   netIPnetworkIP,
		Mask: mask,
	}

	logger.Logger.Println("create_server: Generating server UUID")
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
	logger.Logger.Println("create_server: Getting available nodes from flute module")

	listNodeData, err := ToFluteGetNodes()
	if err != nil {
		logger.Logger.Print(err)
		return nil, err
	}

	// TODO : Currently nrNodes is hard coded to 2. Will get from Web UI (Oboe) later.
	var nrNodes = 2

	// TODO : Get leader node's UUID from selected nodes. Currently, leader node's UUID is provided by subnet data.
	// stage 1.1 update nodes info (server_uuid)
	// stage 1.2 insert nodes to server_node table
	var nodes = listNodeData.Data.ListNode
	var nodeUUIDs []string

	if len(nodes) < nrNodes || len(nodes) == 0 {
		errMsg := "create_server: not enough available nodes"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}

	var nodeSelected = 0
	for _, node := range nodes {
		if nodeSelected > nrNodes {
			break
		}

		logger.Logger.Println("create_server: Updating nodes info to flute module")

		err = ToFluteUpdateNode(node, serverUUID)
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

	logger.Logger.Println("create_server: Getting IP address range")
	firstIP, _ := cidr.AddressRange(&ipNet)
	firstIP = cidr.Inc(firstIP)
	lastIP := firstIP

	for i := 0; i < len(nodes)-1; i++ {
		lastIP = cidr.Inc(lastIP)
	}

	go func() {
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
		err = ToCelloCreateDisk(volumeOS, serverUUID)
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
		err = ToCelloCreateDisk(volumeData, serverUUID)
		if err != nil {
			logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + err.Error())
			return
		}

		// stage 3. ToHarpUpdateSubnet (get subnet info -> create dhcpd config -> update_subnet)
		logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Updating subnet info")
		_, err = ToHarpUpdateSubnet(subnetUUID, serverUUID)
		if err != nil {
			logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + " ToHarpUpdateSubnet: " + err.Error())
			return
		}

		var nodeUUIDsStr = ""
		for i, node := range nodes {
			nodeUUIDsStr += node.UUID
			if i != len(nodes)-1 {
				nodeUUIDsStr += ","
			}
		}
		logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + " nodeUUIDsStr: " + nodeUUIDsStr)

		err = ToHarpCreateDHCPDConfig(subnetUUID, nodeUUIDsStr)
		if err != nil {
			logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + " ToHarpCreateDHCPDConfig: " + err.Error())
			return
		}

		// stage 4. node power on
		logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Turning on leader node")
		var i = 1
		for _, node := range nodes {
			if node.UUID == subnet.Data.Subnet.LeaderNodeUUID {
				_, err := ToFluteOnNode(node.PXEMacAddr)
				if err != nil {
					logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": ToFluteOnNode error: " + err.Error())
					return
				}

				logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": ToFluteOnNode: leader MAC Addr: " + node.PXEMacAddr)

				break
			}

			i++
		}

		if i > len(nodes) {
			logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Failed to find leader node")
			return
		}

		// Wait for leader node to turned on
		time.Sleep(time.Second * time.Duration(config.Flute.WaitForLeaderNodeTimeoutSec))

		logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": " + "Turning on compute nodes")
		for _, node := range nodes {
			if node.UUID == subnet.Data.Subnet.LeaderNodeUUID {
				continue
			}

			_, err := ToFluteOnNode(node.PXEMacAddr)
			if err != nil {
				logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": ToFluteOnNode error: " + err.Error())
				return
			}

			logger.Logger.Println("create_server_routine: server_uuid=" + serverUUID + ": ToFluteOnNode: compute MAC Addr: " + node.PXEMacAddr)
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
}

func updateServer(params graphql.ResolveParams) (interface{}, error) {
	// TODO : Update server stages

	return dao.UpdateServer(params.Args)
}

func deleteServer(params graphql.ResolveParams) (interface{}, error) {
	// TODO : Delete server stages

	return dao.DeleteServer(params.Args)
}

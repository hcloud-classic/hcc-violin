package dao

import (
	"errors"
	"github.com/apparentlymart/go-cidr/cidr"
	uuid "github.com/nu7hatch/gouuid"
	"hcc/violin/action/grpc/client"
	"hcc/violin/action/grpc/pb/rpcflute"
	"hcc/violin/action/grpc/pb/rpcharp"
	pb "hcc/violin/action/grpc/pb/rpcviolin"
	"hcc/violin/data"
	"hcc/violin/driver"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
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
		return nil, errors.New(netmask + " is invalid, netmask should be X.X.X.X form")
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

func doGetSubnet(subnetUUID string) (*net.IPNet, *pb.Subnet, error) {
	var ipNet net.IPNet

	subnet, err := client.RC.GetSubnet(subnetUUID)
	if err != nil {
		logger.Logger.Println(err)
		return nil, nil, err
	}

	if len(subnet.ServerUUID) != 0 {
		errMsg := "createServer: Selected subnet (subnetUUID=" + subnetUUID +
			") is used by one of server (serverUUID=" + subnet.ServerUUID + ")"
		logger.Logger.Println(errMsg)
		return nil, nil, errors.New(errMsg)
	}
	logger.Logger.Println("createServer: subnet info: network IP=" + subnet.NetworkIP +
		", netmask=" + subnet.Netmask)

	netIPnetworkIP := net.ParseIP(subnet.NetworkIP).To4()
	if netIPnetworkIP == nil {
		errMsg := "createServer: got wrong network IP"
		logger.Logger.Println(errMsg)
		return nil, nil, errors.New(errMsg)
	}

	mask, err := checkNetmask(subnet.Netmask)
	if err != nil {
		errMsg := "createServer: got wrong subnet mask"
		logger.Logger.Println(errMsg)
		return nil, nil, errors.New(errMsg)
	}

	ipNet = net.IPNet{
		IP:   netIPnetworkIP,
		Mask: mask,
	}

	return &ipNet, subnet, nil
}

func doGenerateServerUUID() (string, error) {
	out, err := uuid.NewV4()
	if err != nil {
		logger.Logger.Println(err)
		return "", err
	}

	serverUUID := out.String()

	return serverUUID, nil
}

func nodeScheduler(userquota *pb.Quota) ([]string, error) {
	allNodeData, err := driver.SchedulingNodes(userquota)
	// testqwe := allNodeData.(data.ScheduledNodeData).Data.ScheduledNode
	testqwe := allNodeData.(data.ScheduledNodeData)
	if err != nil {
		return nil, err
	}

	// for index := 0; index < len(testqwe.Data.ScheduledNode.NodeList); index++ {
	// 	fmt.Println("++++>", testqwe.Data.ScheduledNode.NodeList[index])
	// }

	//Debug for selected node mutation

	// fmt.Println(testqwe.Data.NodeList)
	// for index := 0; index < len(testqwe.Data.NodeList); index++ {
	// 	fmt.Println(testqwe.Data.NodeList[index])
	// }
	return testqwe.Data.ScheduledNode.NodeList, nil
}

func doGetNodes(userQuota *pb.Quota) ([]pb.Node, error) {
	logger.Logger.Println("doGetNodes(): [Violin Scheduler] Start Scheduling")
	nodeUUIDs, err := nodeScheduler(userQuota)
	logger.Logger.Println("doGetNodes(): [Violin Scheduler] End Scheduling")
	if err != nil {
		return nil, err
	}
	// TODO : Currently nrNodes is hard coded to 2. Will get from Web UI (Oboe) later.
	var nrNodes = userQuota.NumberOfNodes

	if len(nodeUUIDs) < int(nrNodes) || len(nodeUUIDs) == 0 {
		errMsg := "doGetNodes(): not enough available nodes"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}
	var GatherSelectedNodes []pb.Node
	var nodeSelected = 0
	// fmt.Println("Nodes:   ", nodes)
	for _, nodeUUID := range nodeUUIDs {
		if nodeSelected > int(nrNodes) {
			break
		}

		logger.Logger.Println("doGetNodes(): Updating nodes info to flute module")

		eachSelectedNode, err := client.RC.GetNode(nodeUUID)
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}
		GatherSelectedNodes = append(GatherSelectedNodes, pb.Node{
			UUID:        eachSelectedNode.UUID,
			ServerUUID:  eachSelectedNode.ServerUUID,
			BmcMacAddr:  eachSelectedNode.BmcMacAddr,
			BmcIP:       eachSelectedNode.BmcIP,
			PXEMacAddr:  eachSelectedNode.PXEMacAddr,
			Status:      eachSelectedNode.Status,
			CPUCores:    eachSelectedNode.CPUCores,
			Memory:      eachSelectedNode.Memory,
			Description: eachSelectedNode.Description,
			CreatedAt:   eachSelectedNode.CreatedAt,
			Active:      eachSelectedNode.Active,
		})
		// fmt.Println("GatherSelectedNodes\n", GatherSelectedNodes)
		_, err = client.RC.UpdateNode(&rpcflute.ReqUpdateNode{Node: &pb.Node{
			UUID:       eachSelectedNode.UUID,
			ServerUUID: userQuota.ServerUUID,
		}})
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}

		_, errCode, errStr := CreateServerNode(&pb.ReqCreateServerNode{
			ServerNode: &pb.ServerNode{
				ServerUUID: userQuota.ServerUUID,
				NodeUUID:   nodeUUID,
			},
		})
		if errCode != 0 {
			logger.Logger.Println(errStr)
			return nil, errors.New(errStr)
		}

		nodeSelected++
	}

	return GatherSelectedNodes, nil
}

func doGetIPRange(serverSubnet *net.IPNet, nodes []pb.Node) (net.IP, net.IP) {
	firstIP, _ := cidr.AddressRange(serverSubnet)
	firstIP = cidr.Inc(firstIP)
	lastIP := firstIP

	for i := 0; i < len(nodes)-1; i++ {
		lastIP = cidr.Inc(lastIP)
	}

	return firstIP, lastIP
}

func doCreateVolume(serverUUID string, celloParams map[string]interface{}, useType string, firstIP net.IP, gateway string) error {
	userUUID := celloParams["user_uuid"].(string)
	os := celloParams["os"].(string)
	diskSize, err := strconv.Atoi(celloParams["disk_size"].(string))
	if err != nil {
		return err
	}

	var volume model.Volume
	var size int

	switch useType {
	case "os":
		size = model.OSDiskSize
		break
	case "data":
		size = diskSize
		break
	default:
		return errors.New("got invalid useType")
	}

	volume = model.Volume{
		Size:       size,
		Filesystem: os,
		ServerUUID: serverUUID,
		UseType:    useType,
		UserUUID:   userUUID,
		NetworkIP:  firstIP.String(),
		GatewayIP:  gateway,
	}
	logger.Logger.Println("GatewayIP [", gateway, "]")
	err = driver.CreateDisk(volume, serverUUID)
	if err != nil {
		logger.Logger.Println("doCreateVolume: server_uuid=" + serverUUID + ": " + err.Error())
		return err
	}

	return nil
}

func doUpdateSubnet(subnetUUID string, leaderNodeUUID string, serverUUID string) error {
	err := client.RC.UpdateSubnet(&rpcharp.ReqUpdateSubnet{
		Subnet: &rpcharp.Subnet{
			UUID:           subnetUUID,
			LeaderNodeUUID: leaderNodeUUID,
			ServerUUID:     serverUUID,
		},
	})
	if err != nil {
		logger.Logger.Println("doUpdateSubnet: server_uuid=" + serverUUID + " UpdateSubnet: " + err.Error())
		return err
	}

	return nil
}

func doCreateDHCPDConfig(subnetUUID string, serverUUID string, nodes []pb.Node) error {
	var nodeUUIDsStr = ""
	for i := range nodes {
		nodeUUIDsStr += nodes[i].UUID
		if i != len(nodes)-1 {
			nodeUUIDsStr += ","
		}
	}
	logger.Logger.Println("doCreateDHCPDConfig: server_uuid=" + serverUUID + " nodeUUIDsStr: " + nodeUUIDsStr)

	err := client.RC.CreateDHCPDConfig(subnetUUID, nodeUUIDsStr)
	if err != nil {
		logger.Logger.Println("doCreateDHCPDConfig: server_uuid=" + serverUUID + " CreateDHCPDConfig: " + err.Error())
		return err
	}

	return nil
}

func doTurnOnNodes(serverUUID string, leaderNodeUUID string, nodes []pb.Node) error {
	printLogCreateServerRoutine(serverUUID, "Turning on leader node")
	var i = 1
	for i := range nodes {
		if nodes[i].UUID == leaderNodeUUID {
			err := client.RC.OnNode(nodes[i].UUID)
			if err != nil {
				logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
				return err
			}

			logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode: leaderNodeUUID: " + nodes[i].UUID)
			break
		}

		i++
	}

	if i > len(nodes) {
		logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": " + "Failed to find leader node")
		return errors.New("failed to find leader node")
	}

	// Wait for leader node to turned on
	time.Sleep(time.Second * time.Duration(config.Flute.WaitForLeaderNodeTimeoutSec))

	logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": " + "Turning on compute nodes")
	for i := range nodes {
		if nodes[i].UUID == leaderNodeUUID {
			continue
		}

		err := client.RC.OnNode(nodes[i].UUID)
		if err != nil {
			logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
			return err
		}

		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": OnNode: computeNodeUUID: " + nodes[i].UUID)
	}

	return nil
}

func printLogCreateServerRoutine(serverUUID string, msg string) {
	logger.Logger.Println("createServerRoutine(): server_uuid=" + serverUUID + ": " + msg)
}

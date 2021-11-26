package daoext

import (
	"errors"
	"fmt"
	"hcc/violin/action/grpc/client"
	"hcc/violin/data"
	"hcc/violin/driver"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/lib/mysql"
	"hcc/violin/model"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"

	"github.com/apparentlymart/go-cidr/cidr"
	gouuid "github.com/nu7hatch/gouuid"
)

func checkNetmask(netmask string) (net.IPMask, error) {
	var err error

	maskPartsStr := strings.Split(netmask, ".")
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

	mask := net.IPv4Mask(
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

// DoGetSubnet : Get the subnet infos
func DoGetSubnet(subnetUUID string, isUpdate bool) (*net.IPNet, *pb.Subnet, error) {
	var ipNet net.IPNet

	subnet, err := client.RC.GetSubnet(subnetUUID)
	if err != nil {
		logger.Logger.Println(err)
		return nil, nil, err
	}

	if !isUpdate && len(subnet.ServerUUID) != 0 {
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

// DoGenerateServerUUID : Generate a UUID for the server
func DoGenerateServerUUID() (string, error) {
	out, err := gouuid.NewV4()
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

	// Debug for selected node mutation

	// fmt.Println(testqwe.Data.NodeList)
	// for index := 0; index < len(testqwe.Data.NodeList); index++ {
	// 	fmt.Println(testqwe.Data.NodeList[index])
	// }
	return testqwe.Data.ScheduledNode.NodeList, nil
}

// ReadServerNodeList : Get list of server nodes with provided server UUID
func ReadServerNodeList(in *pb.ReqGetServerNodeList) (*pb.ResGetServerNodeList, uint64, string) {
	serverUUID := in.GetServerUUID()
	serverUUIDOk := len(serverUUID) != 0
	if !serverUUIDOk {
		return nil, hcc_errors.ViolinGrpcArgumentError, "ReadServerNodeList(): need a serverUUID argument"
	}

	var serverNodeList pb.ResGetServerNodeList
	var serverNodes []pb.ServerNode
	var pserverNodes []*pb.ServerNode

	var uuid string
	var nodeUUID string
	var createdAt time.Time

	sql := "select * from server_node where server_uuid = ?"

	stmt, err := mysql.Query(sql, serverUUID)
	if err != nil {
		errStr := "ReadServerNodeList(): " + err.Error()
		logger.Logger.Println(errStr)
		return nil, hcc_errors.ViolinSQLOperationFail, errStr
	}
	defer func() {
		_ = stmt.Close()
	}()

	for stmt.Next() {
		err := stmt.Scan(&uuid, &serverUUID, &nodeUUID, &createdAt)
		if err != nil {
			errStr := "ReadServerNodeList(): " + err.Error()
			logger.Logger.Println(errStr)
			if strings.Contains(err.Error(), "no rows in result set") {
				return nil, hcc_errors.ViolinSQLNoResult, errStr
			}
			return nil, hcc_errors.ViolinSQLOperationFail, errStr
		}

		serverNodes = append(serverNodes, pb.ServerNode{
			UUID:       uuid,
			ServerUUID: serverUUID,
			NodeUUID:   nodeUUID,
			CreatedAt:  timestamppb.New(createdAt),
		})
	}

	for i := range serverNodes {
		pserverNodes = append(pserverNodes, &serverNodes[i])
	}

	serverNodeList.ServerNode = pserverNodes

	return &serverNodeList, 0, ""
}

// CheckCreateServerNodeArgs : Check arguments of creating a server node
func CheckCreateServerNodeArgs(reqServerNode *pb.ServerNode) bool {
	serverUUIDOk := len(reqServerNode.ServerUUID) != 0
	nodeUUIDOk := len(reqServerNode.NodeUUID) != 0
	return !(serverUUIDOk && nodeUUIDOk)
}

// DoGetNodes : Get scheduled nodes
func DoGetNodes(userQuota *pb.Quota) ([]pb.Node, error) {
	var reqScheduleServer pb.ReqScheduleHandler
	var reqServer pb.Server
	var reqServerStruct pb.ScheduleServer

	reqScheduleServer.Server = &reqServerStruct
	reqScheduleServer.Server.ScheduleServer = &reqServer
	reqScheduleServer.Server.NumOfNodes = userQuota.GetNumberOfNodes()
	reqScheduleServer.Server.ScheduleServer.CPU = userQuota.GetCPU()
	reqScheduleServer.Server.ScheduleServer.Memory = userQuota.GetMemory()
	reqScheduleServer.Server.ScheduleServer.UUID = userQuota.GetServerUUID()

	logger.Logger.Println("doGetNodes(): [Violin Scheduler] Start Scheduling")
	resNodes, err := client.RC.ScheduleHandler(&reqScheduleServer)
	logger.Logger.Println("doGetNodes(): [Violin Scheduler] End Scheduling")
	if err != nil {
		return nil, err
	}

	nrNodes := userQuota.NumberOfNodes
	retNodes := resNodes.GetNodes()
	nodeList := retNodes.GetShceduledNode()

	fmt.Println("nrNodes : ", nrNodes, "\retNodes:   ", retNodes, "\nnodeList: ", nodeList)

	// if len(nodeList.ShceduledNode) < int(nrNodes) || len(nodeList.ShceduledNode) == 0 {
	if len(nodeList) == 0 {
		errMsg := "doGetNodes(): not enough available nodes"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}
	var GatherSelectedNodes []pb.Node
	nodeSelected := 0

	for _, nodes := range nodeList {
		if nodes.UUID == "" {
			continue
		}

		if nodeSelected > int(nrNodes) {
			break
		}
		logger.Logger.Println("doGetNodes(): Updating nodes info to flute module")

		eachSelectedNode, err := client.RC.GetNode(nodes.GetUUID())
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}
		fmt.Println("eachSelectedNode => ", eachSelectedNode)
		node, err := client.RC.UpdateNode(&pb.ReqUpdateNode{Node: &pb.Node{
			UUID:       eachSelectedNode.UUID,
			Active:     1,
			ServerUUID: userQuota.GetServerUUID(),
		}})
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}

		// fmt.Println("GatherSelectedNodes\n", GatherSelectedNodes)
		GatherSelectedNodes = append(GatherSelectedNodes, pb.Node{
			UUID:        node.UUID,
			ServerUUID:  userQuota.GetServerUUID(),
			BmcMacAddr:  node.BmcMacAddr,
			BmcIP:       node.BmcIP,
			PXEMacAddr:  node.PXEMacAddr,
			Status:      node.Status,
			CPUCores:    node.CPUCores,
			Memory:      node.Memory,
			Description: node.Description,
			CreatedAt:   node.CreatedAt,
			Active:      eachSelectedNode.Active,
			GroupID:     eachSelectedNode.GroupID,
		})

		_, errCode, errStr := CreateServerNode(&pb.ReqCreateServerNode{
			ServerNode: &pb.ServerNode{
				ServerUUID: userQuota.GetServerUUID(),
				NodeUUID:   nodes.GetUUID(),
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

// DoGetIPRange : Get IP range of the subnet
func DoGetIPRange(serverSubnet *net.IPNet, nodes []pb.Node) (net.IP, net.IP) {
	firstIP, _ := cidr.AddressRange(serverSubnet)
	firstIP = cidr.Inc(firstIP)
	lastIP := firstIP

	for i := 0; i < len(nodes)-1; i++ {
		lastIP = cidr.Inc(lastIP)
	}

	return firstIP, lastIP
}

// DoDeleteVolume : Delete the volume
func DoDeleteVolume(serverUUID string) error {
	// userUUID := celloParams["user_uuid"].(string)
	// diskSize, err := strconv.Atoi(celloParams["disk_size"].(string))
	// if err != nil {
	// 	return err
	// }

	var reqDeleteVolume pb.ReqVolumeHandler
	var reqVolume pb.Volume

	reqDeleteVolume.Volume = &reqVolume

	reqDeleteVolume.Volume.ServerUUID = serverUUID
	reqDeleteVolume.Volume.UseType = "os"
	reqDeleteVolume.Volume.Action = "delete"

	logger.Logger.Println("[doDeleteVolume] : ", reqDeleteVolume.Volume)
	resDeleteVolume, err := client.RC.Volhandler(&reqDeleteVolume)
	if err != nil {
		logger.Logger.Println("doDeleteVolume: server_uuid="+serverUUID+": "+err.Error(), "resCreateVolume : ", resDeleteVolume)
		return err
	}

	return nil
}

// DoCreateVolume : Create volumes of os and data
func DoCreateVolume(serverUUID string, celloParams map[string]interface{}, useType string, firstIP net.IP, gateway string) error {
	userUUID := celloParams["user_uuid"].(string)
	diskSize, err := strconv.Atoi(celloParams["disk_size"].(string))
	if err != nil {
		return err
	}

	var reqCreateVolume pb.ReqVolumeHandler
	var reqVolume pb.Volume

	reqCreateVolume.Volume = &reqVolume

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

	reqCreateVolume.Volume.ServerUUID = serverUUID
	reqCreateVolume.Volume.Filesystem = celloParams["os"].(string)
	strSize := strconv.Itoa(size)
	reqCreateVolume.Volume.Size = strSize
	reqCreateVolume.Volume.UserUUID = userUUID
	reqCreateVolume.Volume.UseType = useType
	reqCreateVolume.Volume.Network_IP = firstIP.String()
	reqCreateVolume.Volume.Gateway_IP = gateway

	reqCreateVolume.Volume.Action = "create"
	reqCreateVolume.Volume.GroupID = int64(celloParams["group_id"].(float64))
	logger.Logger.Println("[doCreateVolume] : ", reqCreateVolume.Volume)
	resCreateVolume, err := client.RC.Volhandler(&reqCreateVolume)
	if err != nil {
		logger.Logger.Println("doCreateVolume: server_uuid="+serverUUID+": "+err.Error(), "resCreateVolume : ", resCreateVolume)
		return err
	}

	return nil
}

// DoUpdateSubnet : Update the subnet's leader node UUID and server UUID infos
func DoUpdateSubnet(subnetUUID string, leaderNodeUUID string, serverUUID string) error {
	err := client.RC.UpdateSubnet(&pb.ReqUpdateSubnet{
		Subnet: &pb.Subnet{
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

// DoCreateDHCPDConfig : Create a DHCPD config file for the server
func DoCreateDHCPDConfig(subnetUUID string, serverUUID string) error {
	logger.Logger.Println("doCreateDHCPDConfig: server_uuid=" + serverUUID)

	err := client.RC.CreateDHCPDConfig(subnetUUID)
	if err != nil {
		logger.Logger.Println("doCreateDHCPDConfig: server_uuid=" + serverUUID + " CreateDHCPDConfig: " + err.Error())
		return err
	}

	return nil
}

// DoTurnOnNodes : Turn on selected nodes
func DoTurnOnNodes(serverUUID string, leaderNodeUUID string, nodes []pb.Node) error {
	printLogCreateServerRoutine(serverUUID, "Turning on leader node")

	foundLeaderNode := false

	for i := range nodes {
		if nodes[i].UUID == leaderNodeUUID {
			foundLeaderNode = true

			var err error

			for i := 0; i < int(config.Flute.TurnOnNodesRetryCounts); i++ {
				err = client.RC.OnNode(nodes[i].UUID)
				if err != nil {
					logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
					logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": Retrying for node: " +
						nodes[i].UUID + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(int(config.Flute.TurnOnNodesRetryCounts)))
				} else {
					break
				}
			}

			if err != nil {
				return err
			}

			logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode: leaderNodeUUID: " + nodes[i].UUID)
			break
		}
	}

	if !foundLeaderNode {
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

		logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode: computeNodeUUID: " + nodes[i].UUID)
	}

	return nil
}

// DoTurnOffNodes : Turn off selected nodes
func DoTurnOffNodes(serverUUID string, nodes []pb.Node) error {
	printLogCreateServerRoutine(serverUUID, "Turning off nodes")

	logger.Logger.Println("doTurnOffNodes: server_uuid=" + serverUUID + ": " + "Turning off all of nodes")

	var wait sync.WaitGroup
	var errStr string

	wait.Add(len(nodes))

	for i := range nodes {
		go func(routineServerUUID string, nodeUUID string, routineErrStr string) {
			var err error
			var turnOffErrStr string

			for i := 0; i < int(config.Flute.TurnOffNodesRetryCounts); i++ {
				err = client.RC.OffNode(nodeUUID, true)
				if err != nil {
					turnOffErrStr = "doTurnOffNodes: server_uuid=" + routineServerUUID + ": OffNode error: " + err.Error()
					logger.Logger.Println(turnOffErrStr)
					logger.Logger.Println("doTurnOffNodes: server_uuid=" + routineServerUUID + ": Retrying for node: " +
						nodeUUID + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(int(config.Flute.TurnOffNodesRetryCounts)))
				} else {
					break
				}
			}

			if err != nil {
				routineErrStr += turnOffErrStr + "\n"
			}

			logger.Logger.Println("doTurnOffNodes: server_uuid=" + routineServerUUID + ": OffNode: NodeUUID: " + nodeUUID)

			wait.Done()
		}(serverUUID, nodes[i].UUID, errStr)
	}

	wait.Wait()

	if errStr != "" {
		return errors.New(errStr)
	}

	return nil
}

func printLogCreateServerRoutine(serverUUID string, msg string) {
	logger.Logger.Println("createServerRoutine(): server_uuid=" + serverUUID + ": " + msg)
}

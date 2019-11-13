package driver

import (
	"errors"
	"hcc/violin/action/rabbitmq"
	"hcc/violin/dao"
	"hcc/violin/data"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/graphql-go/graphql"
	uuid "github.com/nu7hatch/gouuid"
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

func doGetSubnet(subnetUUID string) (net.IPNet, model.Subnet, error) {
	var ipNet net.IPNet
	var subnet model.Subnet

	subnet, err := GetSubnet(subnetUUID)
	if err != nil {
		logger.Logger.Println(err)
		return ipNet, subnet, err
	}

	if len(subnet.ServerUUID) != 0 {
		errMsg := "createServer: Selected subnet (subnetUUID=" + subnetUUID +
			") is used by one of server (serverUUID=" + subnet.ServerUUID + ")"
		logger.Logger.Println(errMsg)
		return ipNet, subnet, errors.New(errMsg)
	}
	logger.Logger.Println("createServer: subnet info: network IP=" + subnet.NetworkIP +
		", netmask=" + subnet.Netmask)

	netIPnetworkIP := net.ParseIP(subnet.NetworkIP).To4()
	if netIPnetworkIP == nil {
		errMsg := "createServer: got wrong network IP"
		logger.Logger.Println(errMsg)
		return ipNet, subnet, errors.New(errMsg)
	}

	mask, err := checkNetmask(subnet.Netmask)
	if err != nil {
		errMsg := "createServer: got wrong subnet mask"
		logger.Logger.Println(errMsg)
		return ipNet, subnet, errors.New(errMsg)
	}

	ipNet = net.IPNet{
		IP:   netIPnetworkIP,
		Mask: mask,
	}

	return ipNet, subnet, nil
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

func doGetNodes(serverUUID string) ([]model.Node, error) {
	listNodeData, err := GetNodes()
	nodes := listNodeData.(data.ListNodeData).Data.ListNode
	if err != nil {
		logger.Logger.Print(err)
		return nil, err
	}

	// TODO : Currently nrNodes is hard coded to 2. Will get from Web UI (Oboe) later.
	var nrNodes = 2

	// TODO : Get leader node's UUID from selected nodes. Currently, leader node's UUID is provided by subnet data.
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

	return nodes, nil
}

func doGetIPRange(serverSubnet net.IPNet, nodes []model.Node) (net.IP, net.IP) {
	firstIP, _ := cidr.AddressRange(&serverSubnet)
	firstIP = cidr.Inc(firstIP)
	lastIP := firstIP

	for i := 0; i < len(nodes)-1; i++ {
		lastIP = cidr.Inc(lastIP)
	}

	return firstIP, lastIP
}

func doCreateVolume(serverUUID string, params graphql.ResolveParams, useType string, firstIP net.IP) error {
	userUUID := params.Args["user_uuid"].(string)
	os := params.Args["os"].(string)
	diskSize := params.Args["disk_size"].(int)

	var volume model.Volume

	switch useType {
	case "os":
		volume.Size = model.OSDiskSize
		break
	case "data":
		volume.Size = diskSize
		break
	default:
		return errors.New("got invalid useType")
	}

	volume = model.Volume{
		Filesystem: os,
		ServerUUID: serverUUID,
		UseType:    useType,
		UserUUID:   userUUID,
		NetworkIP:  firstIP.String(),
	}

	err := CreateDisk(volume, serverUUID)
	if err != nil {
		logger.Logger.Println("doCreateVolume: server_uuid=" + serverUUID + ": " + err.Error())
		return err
	}

	return nil
}

func doUpdateSubnet(subnetUUID string, serverUUID string) error {
	_, err := UpdateSubnet(subnetUUID, serverUUID)
	if err != nil {
		logger.Logger.Println("doUpdateSubnet: server_uuid=" + serverUUID + " UpdateSubnet: " + err.Error())
		return err
	}

	return nil
}

func doCreateDHCPDConfig(subnetUUID string, serverUUID string, nodes []model.Node) error {
	var nodeUUIDsStr = ""
	for i, node := range nodes {
		nodeUUIDsStr += node.UUID
		if i != len(nodes)-1 {
			nodeUUIDsStr += ","
		}
	}
	logger.Logger.Println("doCreateDHCPDConfig: server_uuid=" + serverUUID + " nodeUUIDsStr: " + nodeUUIDsStr)

	err := CreateDHCPDConfig(subnetUUID, nodeUUIDsStr)
	if err != nil {
		logger.Logger.Println("doCreateDHCPDConfig: server_uuid=" + serverUUID + " CreateDHCPDConfig: " + err.Error())
		return err
	}

	return nil
}

func doTurnOnNodes(serverUUID string, leaderNodeUUID string, nodes []model.Node) error {
	printLogCreateServerRoutine(serverUUID, "Turning on leader node")
	var i = 1
	for _, node := range nodes {
		if node.UUID == leaderNodeUUID {
			_, err := OnNode(node.PXEMacAddr)
			if err != nil {
				logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
				return err
			}

			logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode: leader MAC Addr: " + node.PXEMacAddr)
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
	for _, node := range nodes {
		if node.UUID == leaderNodeUUID {
			continue
		}

		_, err := OnNode(node.PXEMacAddr)
		if err != nil {
			logger.Logger.Println("doTurnOnNodes: server_uuid=" + serverUUID + ": OnNode error: " + err.Error())
			return err
		}

		logger.Logger.Println("createServer_routine: server_uuid=" + serverUUID + ": OnNode: compute MAC Addr: " + node.PXEMacAddr)
	}

	return nil
}

func doHccCLI(serverUUID string, firstIP net.IP, lastIP net.IP) error {
	logger.Logger.Println("doHccCLI: server_uuid=" + serverUUID + ": " + "Preparing controlAction")

	hccaction := model.HccAction{

		ActionArea:  "nodes",
		ActionClass: "add",
		ActionScope: "0",
		HccIPRange:  firstIP.String() + " " + lastIP.String(),
		ServerUUID:  serverUUID,
	}

	hcctype := model.Action{
		ActionType: "hcc",
		HccType:    hccaction,
	}

	controlAction := model.Control{
		Publisher: "violin",
		Receiver:  "violin",
		Control:   hcctype,
	}

	err := rabbitmq.ViolinToViola(controlAction)
	if err != nil {
		logger.Logger.Println("doHccCLI: server_uuid=" + serverUUID + ": " + err.Error())
		return err
	}

	return nil
}

func printLogCreateServerRoutine(serverUUID string, msg string) {
	logger.Logger.Println("createServerRoutine: server_uuid=" + serverUUID + ": " + msg)
}

// CreateServer : Do server creation works
func CreateServer(params graphql.ResolveParams) (interface{}, error) {
	subnetUUID := params.Args["subnet_uuid"].(string)

	logger.Logger.Println("createServer: Getting subnet info from harp module")
	serverSubnet, subnet, err := doGetSubnet(subnetUUID)
	if err != nil {
		return nil, err
	}

	logger.Logger.Println("createServer: Generating server UUID")
	serverUUID, err := doGenerateServerUUID()
	if err != nil {
		return nil, err
	}

	logger.Logger.Println("createServer: Getting available nodes from flute module")
	nodes, err := doGetNodes(serverUUID)
	if err != nil {
		return nil, err
	}

	logger.Logger.Println("createServer: Getting IP address range")
	firstIP, lastIP := doGetIPRange(serverSubnet, nodes)

	go func(routineServerUUID string, routineSubnet model.Subnet, routineNodes []model.Node,
		routineParams graphql.ResolveParams, routineFirstIP net.IP, routineLastIP net.IP) {
		var routineError error

		printLogCreateServerRoutine(routineServerUUID, "Creating os volume")
		routineError = doCreateVolume(routineServerUUID, routineParams, "os", routineFirstIP)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Creating data volume")
		routineError = doCreateVolume(routineServerUUID, routineParams, "data", routineFirstIP)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Updating subnet info")
		routineError = doUpdateSubnet(routineSubnet.UUID, routineServerUUID)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Creating DHCPD config file")
		routineError = doCreateDHCPDConfig(routineSubnet.UUID, routineServerUUID, routineNodes)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Turning on nodes")
		routineError = doTurnOnNodes(routineServerUUID, routineSubnet.LeaderNodeUUID, routineNodes)
		if routineError != nil {
			goto ERROR
		}

		printLogCreateServerRoutine(routineServerUUID, "Preparing controlAction")

		printLogCreateServerRoutine(routineServerUUID, "Running Hcc CLI")
		routineError = doHccCLI(routineServerUUID, routineFirstIP, routineLastIP)
		if routineError != nil {
			goto ERROR
		}
		// while checking Cello DB cluster status is runnig in N times, until retry is expired

		return

	ERROR:
		printLogCreateServerRoutine(routineServerUUID, routineError.Error())
	}(serverUUID, subnet, nodes, params, firstIP, lastIP)

	return dao.CreateServer(serverUUID, params.Args)
}

// UpdateServer : Do server updating works
func UpdateServer(params graphql.ResolveParams) (interface{}, error) {
	// TODO : Update server stages

	return dao.UpdateServer(params.Args)
}

// DeleteServer : Do server deleting works
func DeleteServer(params graphql.ResolveParams) (interface{}, error) {
	// TODO : Delete server stages

	return dao.DeleteServer(params.Args)
}

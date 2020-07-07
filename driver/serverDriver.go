package driver

import (
	"errors"
	"fmt"
	"hcc/violin/dao"
	"hcc/violin/data"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"net"
	"strconv"
	"strings"

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

func NodeScheduler(userquota model.Quota) ([]string, error) {
	allNodeData, err := SchedulingNodes(userquota)
	testqwe := allNodeData.(data.ScheduledNodeData)
	if err != nil {
		fmt.Println(err)
	}

	return testqwe.Data.ScheduledNode.NodeList, nil
}

func doGetNodes(userquota model.Quota) ([]model.Node, error) {
	logger.Logger.Println("[Violin Scheduler] Start Scheduling")
	nodes, err := NodeScheduler(userquota)
	logger.Logger.Println("[Violin Scheduler] End Scheduling")
	if err != nil {
		return nil, err
	}
	var nrNodes = userquota.NumberOfNodes

	var nodeUUIDs []string

	if len(nodes) < nrNodes || len(nodes) == 0 {
		errMsg := "createServer: not enough available nodes"
		logger.Logger.Println(errMsg)
		return nil, errors.New(errMsg)
	}
	var GatherSelectedNodes []model.Node
	var nodeSelected = 0
	for _, node := range nodes {
		if nodeSelected > nrNodes {
			break
		}

		logger.Logger.Println("createServer: Updating nodes info to flute module")

		SingleNodeData, err := GetSingleNode(node)

		eachSelectedNode := SingleNodeData.(data.SingleNodeData).Data.SingleNode
		GatherSelectedNodes = append(GatherSelectedNodes, eachSelectedNode)
		err = UpdateNode(eachSelectedNode, userquota.ServerUUID)
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}

		args := make(map[string]interface{})
		args["server_uuid"] = userquota.ServerUUID
		args["node_uuid"] = node
		_, err = dao.CreateServerNode(args)
		if err != nil {
			logger.Logger.Println(err)
			return nil, err
		}

		nodeUUIDs = append(nodeUUIDs, node)

		nodeSelected++
	}

	return GatherSelectedNodes, nil
}

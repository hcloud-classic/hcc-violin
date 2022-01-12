package rabbitmq

import (
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"net"
)

// HccCLI : Send nodes add action to viola
func HccCLI(serverUUID string, firstIP net.IP, lastIP net.IP, token string, gateway string) error {
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
		Token:     token,
	}

	err := ViolinToViola(controlAction, gateway)
	if err != nil {
		logger.Logger.Println("doHccCLI: server_uuid=" + serverUUID + ": " + err.Error())
		return err
	}

	return nil
}

package rabbitmq

import (
	"encoding/json"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"net"

	"innogrid.com/hcloud-classic/pb"

	"github.com/TylerBrock/colorjson"
	"github.com/streadway/amqp"
)

// ViolinToViola : Publish 'run_hcc_cli' queues to RabbitMQ channel
func ViolinToViola(action model.Control, gateway string) error {
	qCreate, err := Channel.QueueDeclare(
		gateway+"_to_viola",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("ViolinToViola(" + gateway + "_to_viola) : Failed to declare a create queue")
		return err
	}

	body, _ := json.Marshal(action)
	err = Channel.Publish(
		"",
		qCreate.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			Body:            body,
		})
	var obj map[string]interface{}
	json.Unmarshal([]byte(body), &obj)

	// Make a custom formatter with indent set
	f := colorjson.NewFormatter()
	f.Indent = 4

	// Marshall the Colorized JSON
	s, _ := f.Marshal(obj)
	// fmt.Println(string(s))
	logger.Logger.Println("doHcc Action [", string(s), "]")

	if err != nil {
		logger.Logger.Println("ViolinToViola(" + gateway + "_to_viola) : Failed to register publisher")
		return err
	}

	return nil
}

// QueueCreateServer : Publish server creating queues to RabbitMQ channel
func QueueCreateServer(routineServerUUID string, routineOS string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
	celloParams map[string]interface{}, routineFirstIP net.IP, routineLastIP net.IP, token string) error {
	qCreate, err := Channel.QueueDeclare(
		"create_server",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("QueueCreateServer: Failed to declare a create queue")
		return err
	}

	body, _ := json.Marshal(
		createServerDataStruct{
			RoutineServerUUID: routineServerUUID,
			RoutineServerOS:   routineOS,
			RoutineSubnet: pb.Subnet{
				UUID:           routineSubnet.UUID,
				NetworkIP:      routineSubnet.NetworkIP,
				Netmask:        routineSubnet.Netmask,
				Gateway:        routineSubnet.Gateway,
				NextServer:     routineSubnet.NextServer,
				NameServer:     routineSubnet.NameServer,
				DomainName:     routineSubnet.DomainName,
				LeaderNodeUUID: routineSubnet.LeaderNodeUUID,
				OS:             routineOS,
				SubnetName:     routineSubnet.SubnetName,
				CreatedAt:      routineSubnet.CreatedAt,
			},
			RoutineNodes:   routineNodes,
			CelloParams:    celloParams,
			RoutineFirstIP: routineFirstIP,
			RoutineLastIP:  routineLastIP,
			Token:          token,
			Action:         "create",
		})
	err = Channel.Publish(
		"",
		qCreate.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			Body:            body,
		})
	if err != nil {
		logger.Logger.Println("QueueCreateServer: Failed to register publisher")
		return err
	}

	return nil
}

// QueueUpdateServerNodes : Publish server updating queues to RabbitMQ channel
func QueueUpdateServerNodes(routineServerUUID string, routineSubnet *pb.Subnet, routineNodes []pb.Node,
	routineFirstIP net.IP, routineLastIP net.IP, token string) error {
	qCreate, err := Channel.QueueDeclare(
		"create_server",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("QueueUpdateServerNodes: Failed to declare a update queue")
		return err
	}

	body, _ := json.Marshal(
		createServerDataStruct{
			RoutineServerUUID: routineServerUUID,
			RoutineSubnet: pb.Subnet{
				UUID:           routineSubnet.UUID,
				GroupID:        routineSubnet.GroupID,
				NetworkIP:      routineSubnet.NetworkIP,
				Netmask:        routineSubnet.Netmask,
				Gateway:        routineSubnet.Gateway,
				NextServer:     routineSubnet.NextServer,
				NameServer:     routineSubnet.NameServer,
				DomainName:     routineSubnet.DomainName,
				LeaderNodeUUID: routineSubnet.LeaderNodeUUID,
				OS:             routineSubnet.OS,
				SubnetName:     routineSubnet.SubnetName,
				CreatedAt:      routineSubnet.CreatedAt,
			},
			RoutineNodes:   routineNodes,
			RoutineFirstIP: routineFirstIP,
			RoutineLastIP:  routineLastIP,
			Token:          token,
			Action:         "update",
		})
	err = Channel.Publish(
		"",
		qCreate.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			Body:            body,
		})
	if err != nil {
		logger.Logger.Println("QueueUpdateServerNodes: Failed to register publisher")
		return err
	}

	return nil
}

// QueueDeleteServer : Publish server deleting queues to RabbitMQ channel
func QueueDeleteServer(routineServerUUID string, token string) error {
	qCreate, err := Channel.QueueDeclare(
		"create_server",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("QueueDeleteServer: Failed to declare a update queue")
		return err
	}

	body, _ := json.Marshal(
		createServerDataStruct{
			RoutineServerUUID: routineServerUUID,
			Token:             token,
			Action:            "delete",
		})
	err = Channel.Publish(
		"",
		qCreate.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:     "text/plain",
			ContentEncoding: "utf-8",
			Body:            body,
		})
	if err != nil {
		logger.Logger.Println("QueueDeleteServer: Failed to register publisher")
		return err
	}

	return nil
}

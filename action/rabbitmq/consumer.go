package rabbitmq

import (
	"encoding/json"
	"hcc/violin/dao"
	"hcc/violin/lib/logger"
	"hcc/violin/model"
	"log"
)

// ReturnNodes : Consume 'return_nodes' queues from RabbitMQ channel
func ReturnNodes() error {
	qCreate, err := Channel.QueueDeclare(
		"return_nodes",
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		logger.Logger.Println("return_nodes: Failed to declare a create queue")
		return err
	}

	msgsCreate, err := Channel.Consume(
		qCreate.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logger.Logger.Println("return_nodes: Failed to register consumer")
		return err
	}

	go func() {
		for d := range msgsCreate {
			log.Printf("return_nodes: Received a create message: %s", d.Body)

			var nodes []model.Node
			err = json.Unmarshal(d.Body, &nodes)
			if err != nil {
				logger.Logger.Println("get_nodes: Failed to unmarshal subnet data")
				return
			}

			serverUUID := server.UUID
			nodeNr := server.NodeNr

			nodes, err := dao.GetAvailableNodes()
			if err != nil {
				logger.Logger.Println(err)
				return
			}

			if nodeNr > len(nodes) {
				logger.Logger.Println("get_nodes: Requested nodeNr is lager than available nodes count")
				return
			}

			for i, node := range nodes {
				if i > nodeNr {
					break
				}
				err := dao.UpdateNodeServerUUID(node, serverUUID)
				if err != nil {
					logger.Logger.Println("get_nodes: error occurred while updating server_uuid of node (UUID = " + node.UUID)
					return
				}
			}

			nodesSelected, err := dao.GetNodesOfServer(serverUUID)
			if err != nil {
				logger.Logger.Println(err)
				return
			}

			err = ReturnNodes(nodesSelected)
			if err != nil {
				logger.Logger.Println(err)
				return
			}

			/*
				TODO
				- 1. select * from node where server_uuid is not null
				- 2. 필요한 갯수 만큼 get
				- 3. get 한 노드들에 대해 update node set server_uuid = [server_uuid]
				- 4. return nodeUUIDs: select * from node where server_uuid = [server_uuid]
				5. publish to harp: create_dhcpd_conf
			*/

			//logger.Logger.Println("create_dhcpd_config: UUID = " + uuid + ": " + result)
		}
	}()

	return nil
}
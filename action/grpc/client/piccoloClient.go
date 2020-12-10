package client

import (
	"context"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/pb/rpcpiccolo"
	pb "hcc/violin/action/grpc/pb/rpcviolin"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"
)

var piccoloConn *grpc.ClientConn

func initPiccolo() error {
	var err error

	addr := config.Piccolo.ServerAddress + ":" + strconv.FormatInt(config.Piccolo.ServerPort, 10)
	piccoloConn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	RC.piccolo = rpcpiccolo.NewPiccoloClient(piccoloConn)
	logger.Logger.Println("gRPC piccolo client ready")

	return nil
}

func closePiccolo() {
	_ = piccoloConn.Close()
}

// WriteServerAction : Write server actions to the sqlite database file
func (rc *RPCClient) WriteServerAction(serverUUID string, action string, result string,
	errStr string, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Piccolo.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	_, err := rc.piccolo.WriteServerAction(ctx, &rpcpiccolo.ReqWriteServerAction{
		ServerUUID: serverUUID,
		ServerAction: &pb.ServerAction{
			Action: action,
			Result: result,
			ErrStr: errStr,
			Token:  token,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

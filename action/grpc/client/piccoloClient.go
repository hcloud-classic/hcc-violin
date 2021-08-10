package client

import (
	"context"
	"google.golang.org/grpc"
	"hcc/violin/action/grpc/errconv"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"innogrid.com/hcloud-classic/hcc_errors"
	"innogrid.com/hcloud-classic/pb"
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

	RC.piccolo = pb.NewPiccoloClient(piccoloConn)
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
	_, err := rc.piccolo.WriteServerAction(ctx, &pb.ReqWriteServerAction{
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

// GetGroupList : Get list of the group
func (rc *RPCClient) GetGroupList(_ *pb.Empty) (*pb.ResGetGroupList, *hcc_errors.HccErrorStack) {
	var errStack *hcc_errors.HccErrorStack

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Piccolo.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetGroupList, err := rc.piccolo.GetGroupList(ctx, &pb.Empty{})
	if err != nil {
		hccErrStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(hcc_errors.ViolinGrpcRequestError, "GetGroupList(): "+err.Error()))
		return nil, hccErrStack
	}
	if es := resGetGroupList.GetHccErrorStack(); es != nil {
		errStack = errconv.GrpcStackToHcc(es)
	}

	return resGetGroupList, errStack
}

// GetQuota : Get the quota of the group
func (rc *RPCClient) GetQuota(groupID int64) (*pb.ResGetQuota, *hcc_errors.HccErrorStack) {
	var errStack *hcc_errors.HccErrorStack

	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Piccolo.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resGetQuota, err := rc.piccolo.GetQuota(ctx, &pb.ReqGetQuota{GroupID: groupID})
	if err != nil {
		hccErrStack := hcc_errors.NewHccErrorStack(hcc_errors.NewHccError(hcc_errors.HarpGrpcRequestError, "GetQuota(): "+err.Error()))
		return nil, hccErrStack
	}
	if es := resGetQuota.GetHccErrorStack(); es != nil {
		errStack = errconv.GrpcStackToHcc(es)
	}

	return resGetQuota, errStack
}

package client

import (
	"context"
	"hcc/violin/action/grpc/pb/rpcviolin_scheduler"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

var schedulerConn *grpc.ClientConn

func initScheduler() error {
	var err error

	addr := config.ViolinScheduler.ServerAddress + ":" + strconv.FormatInt(config.ViolinScheduler.ServerPort, 10)
	schedulerConn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	RC.scheduler = rpcviolin_scheduler.NewSchedulerClient(schedulerConn)
	logger.Logger.Println("gRPC violin client ready")

	return nil
}

func closeScheduler() {
	_ = schedulerConn.Close()
}

// ScheduleHandler : Create a server
func (rc *RPCClient) ScheduleHandler(in *rpcviolin_scheduler.ReqScheduleHandler) (*rpcviolin_scheduler.ResScheduleHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.ViolinScheduler.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resScheduledNode, err := rc.scheduler.ScheduleHandler(ctx, in)
	if err != nil {
		return nil, err
	}

	return resScheduledNode, nil
}

package client

import (
	"context"
	errors2 "errors"
	"github.com/hcloud-classic/pb"
	"hcc/violin/action/grpc/errconv"
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

	RC.scheduler = pb.NewSchedulerClient(schedulerConn)
	logger.Logger.Println("gRPC violin-scheduler client ready")

	return nil
}

func closeScheduler() {
	_ = schedulerConn.Close()
}

// ScheduleHandler : Schedule of getting nodes
func (rc *RPCClient) ScheduleHandler(in *pb.ReqScheduleHandler) (*pb.ResScheduleHandler, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.ViolinScheduler.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resScheduledNode, err := rc.scheduler.ScheduleHandler(ctx, in)
	if err != nil {
		return nil, err
	}

	hccErrStack := errconv.GrpcStackToHcc(resScheduledNode.HccErrorStack)
	errors := *hccErrStack.ConvertReportForm().Stack()
	if len(errors) != 0 && errors[0].Code() != 0 {
		return nil, errors2.New(errors[0].Text())
	}

	return resScheduledNode, nil
}

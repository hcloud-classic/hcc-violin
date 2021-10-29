package client

import (
	"context"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"innogrid.com/hcloud-classic/pb"
)

var pianoConn *grpc.ClientConn

func initPiano() error {
	var err error

	addr := config.Piano.ServerAddress + ":" + strconv.FormatInt(config.Piano.ServerPort, 10)
	pianoConn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	RC.piano = pb.NewPianoClient(pianoConn)
	logger.Logger.Println("gRPC piano client ready")

	return nil
}

func closePiano() {
	_ = pianoConn.Close()
}

func pingPiano() bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(config.Piano.ServerAddress,
		strconv.FormatInt(config.Piano.ServerPort, 10)),
		time.Duration(config.Grpc.ClientPingTimeoutMs)*time.Millisecond)
	if err != nil {
		return false
	}
	if conn != nil {
		defer func() {
			_ = conn.Close()
		}()
		return true
	}

	return false
}

func checkPiano() {
	ticker := time.NewTicker(time.Duration(config.Grpc.ClientPingIntervalMs) * time.Millisecond)
	go func() {
		connOk := true
		for range ticker.C {
			pingOk := pingPiano()
			if pingOk {
				if !connOk {
					logger.Logger.Println("checkPiano(): Ping Ok! Resetting connection...")
					closePiano()
					err := initPiano()
					if err != nil {
						logger.Logger.Println("checkPiano(): " + err.Error())
						continue
					}
					connOk = true
				}
			} else {
				if connOk {
					logger.Logger.Println("checkPiano(): Piano module seems dead. Pinging...")
				}
				connOk = false
			}
		}
	}()
}

// Telegraph : Get the metric data from influxDB
func (rc *RPCClient) Telegraph(in *pb.ReqMetricInfo) (*pb.ResMonitoringData, error) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Piano.RequestTimeoutMs)*time.Millisecond)
	defer cancel()
	resMonitoringData, err := rc.piano.Telegraph(ctx, in)
	if err != nil {
		return nil, err
	}

	return resMonitoringData, nil
}

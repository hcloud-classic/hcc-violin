package autoscale

import (
	"encoding/json"
	"hcc/violin/action/grpc/client"
	"hcc/violin/dao"
	"hcc/violin/lib/config"
	"hcc/violin/lib/logger"
	"innogrid.com/hcloud-classic/pb"
	"strconv"
	"strings"
	"time"
)

var checkServerResourceLocked = false

func delayMillisecond(n time.Duration) {
	time.Sleep(n * time.Millisecond)
}

func checkServerResourceLock() {
	checkServerResourceLocked = true
}

func checkServerResourceUnlock() {
	checkServerResourceLocked = false
}

func getMetricPercent(resMonitoringData *pb.ResMonitoringData) (int, error) {
	var metric metric
	var currentUsage = 0

	err := json.Unmarshal(resMonitoringData.MonitoringData.Result, &metric)
	if err != nil {
		logger.Logger.Println("getCPUPercent(): err=" + err.Error())
		return 0, err
	}

	if len(metric) < 1 {
		logger.Logger.Println("getMetricPercent(): Failed to read metric")
		return 0, err
	}

	if len(metric[0].Series) > 0 {
		if len(metric[0].Series[0].Values) > 0 {
			if len(metric[0].Series[0].Values[0]) == 3 {
				currentUsage = int((metric[0].Series[0].Values[0][1]).(float64))
			} else {
				logger.Logger.Println("getMetricPercent(): Got wrong values while reading metric")
				return 0, err
			}
		} else {
			logger.Logger.Println("getMetricPercent(): Failed to read values in metric")
			return 0, err
		}
	} else {
		logger.Logger.Println("getMetricPercent(): Failed to read series in metric")
		return 0, err
	}

	return currentUsage, nil
}

func getCPUUsagePercent(serverUUID string) (int, error) {
	resMonitoringData, err := client.RC.Telegraph(&pb.ReqMetricInfo{
		MetricInfo: &pb.MetricInfo{
			AggregateFn: ",",
			Uuid:        serverUUID,
			Metric:      "cpu",
			SubMetric:   "usage_user,usage_system",
			OrderBy:     "period",
			Period:      "ms",
			GroupBy:     "cpu",
			Limit:       "10",
		},
	})
	if err != nil {
		logger.Logger.Println("getCPUUsagePercent(): err=" + err.Error())
		return 0, err
	}

	return getMetricPercent(resMonitoringData)
}

func getMemoryUsagePercent(serverUUID string) (int, error) {
	resMonitoringData, err := client.RC.Telegraph(&pb.ReqMetricInfo{
		MetricInfo: &pb.MetricInfo{
			AggregateFn: ",",
			Uuid:        serverUUID,
			Metric:      "mem",
			SubMetric:   "used_percent,available_percent",
			OrderBy:     "period",
			Period:      "ms",
			Limit:       "10",
		},
	})
	if err != nil {
		logger.Logger.Println("getMemoryUsagePercent(): err=" + err.Error())
		return 0, err
	}

	return getMetricPercent(resMonitoringData)
}

func doCheckServerResource() {
	serverList, errCode, errStr := dao.ReadServerList(&pb.ReqGetServerList{})
	if errCode != 0 {
		logger.Logger.Println("doCheckServerResource(): err=" + errStr)
		return
	}

	var reason = "AutoScale Triggered"

	for _, server := range serverList.Server {
		if strings.ToLower(server.Status) == "creating" {
			continue
		}

		cpuUsagePercent, err := getCPUUsagePercent(server.UUID)
		if err != nil {
			logger.Logger.Print("doCheckServerResource(): cpuUsagePercent, errStr=" + err.Error())
			return
		}
		if config.AutoScale.Debug == "on" {
			logger.Logger.Print("doCheckServerResource(): serverUUID=" + server.UUID + ", cpuUsagePercent=" + strconv.Itoa(cpuUsagePercent))
		}

		if cpuUsagePercent >= int(config.AutoScale.AutoScaleTriggerCPUUsagePercent) {
			err = client.RC.WriteServerAlarm(server.UUID, reason,
				"CPU Usage is higher than "+strconv.Itoa(int(config.AutoScale.AutoScaleTriggerCPUUsagePercent))+"%")
			if err != nil {
				logger.Logger.Println("doCheckServerResource(): cpuUsagePercent, errStr=" + err.Error())
			}
		}

		memoryUsagePercent, err := getMemoryUsagePercent(server.UUID)
		if err != nil {
			logger.Logger.Print("doCheckServerResource(): memoryUsagePercent, errStr=" + err.Error())
			return
		}
		if config.AutoScale.Debug == "on" {
			logger.Logger.Print("doCheckServerResource(): serverUUID=" + server.UUID + ", memoryUsagePercent=" + strconv.Itoa(memoryUsagePercent))
		}

		if memoryUsagePercent >= int(config.AutoScale.AutoScaleTriggerMemoryUsagePercent) {
			err = client.RC.WriteServerAlarm(server.UUID, reason,
				"Memory Usage is higher than "+strconv.Itoa(int(config.AutoScale.AutoScaleTriggerMemoryUsagePercent))+"%")
			if err != nil {
				logger.Logger.Println("doCheckServerResource(): memoryUsagePercent, errStr=" + err.Error())
			}
		}
	}

	if config.AutoScale.Debug == "on" {
		logger.Logger.Print("doCheckServerResource(): Done")
	}
}

func queueCheckServerResource() {
	go func() {
		if config.AutoScale.Debug == "on" {
			logger.Logger.Println("queueCheckServerStatus(): Queued of running CheckServerStatus() after " + strconv.Itoa(int(config.AutoScale.CheckServerResourceIntervalMs)) + "ms")
		}
		delayMillisecond(time.Duration(config.AutoScale.CheckServerResourceIntervalMs))
		CheckServerResource()
	}()
}

// CheckServerResource : Check each server's resources to trigger auto-scale.
func CheckServerResource() {
	if checkServerResourceLocked {
		if config.AutoScale.Debug == "on" {
			logger.Logger.Println("CheckServerStatus(): Locked")
		}
		for true {
			if !checkServerResourceLocked {
				break
			}
			if config.AutoScale.Debug == "on" {
				logger.Logger.Println("CheckServerStatus(): Rerun after " + strconv.Itoa(int(config.AutoScale.CheckServerResourceIntervalMs)) + "ms")
			}
			delayMillisecond(time.Duration(config.AutoScale.CheckServerResourceIntervalMs))
		}
	}

	go func() {
		checkServerResourceLock()
		if config.AutoScale.Debug == "on" {
			logger.Logger.Println("CheckServerStatus(): Running UpdateServerStatus()")
		}
		doCheckServerResource()
		checkServerResourceUnlock()
	}()

	queueCheckServerResource()
}

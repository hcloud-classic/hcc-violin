package pid

import (
	"hcc/violin/lib/fileutil"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

var violinPIDFileLocation = "/var/run"
var violinPIDFile = "/var/run/violin.pid"

// IsViolinRunning : Check if violin is running
func IsViolinRunning() (running bool, pid int, err error) {
	if _, err := os.Stat(violinPIDFile); os.IsNotExist(err) {
		return false, 0, nil
	}

	pidStr, err := ioutil.ReadFile(violinPIDFile)
	if err != nil {
		return false, 0, err
	}

	violinPID, _ := strconv.Atoi(string(pidStr))

	proc, err := os.FindProcess(violinPID)
	if err != nil {
		return false, 0, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, violinPID, nil
	}

	return false, 0, nil
}

// WriteViolinPID : Write violin PID to "/var/run/violin.pid"
func WriteViolinPID() error {
	pid := os.Getpid()

	err := fileutil.CreateDirIfNotExist(violinPIDFileLocation)
	if err != nil {
		return err
	}

	err = fileutil.WriteFile(violinPIDFile, strconv.Itoa(pid))
	if err != nil {
		return err
	}

	return nil
}

// DeleteViolinPID : Delete the violin PID file
func DeleteViolinPID() error {
	err := fileutil.DeleteFile(violinPIDFile)
	if err != nil {
		return err
	}

	return nil
}

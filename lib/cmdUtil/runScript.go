package cmdUtil

import (
	"hcc/violin/lib/logger"
	"os/exec"
)

func RunScript(filepath string) error {
	logger.Logger.Println("Running script file: " + filepath)

	cmd := exec.Command("csh", filepath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

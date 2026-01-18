package agent

import (
	"fmt"
	"os/exec"
	"runtime"
)

func ExecuteCommand(command string) (string, error) {
	if runtime.GOOS == "windows" {
		command = fmt.Sprintf("cmd /c %s", command)
	} else {
		command = fmt.Sprintf("/bin/sh -c %s", command)
	}

	output, err := exec.Command(command).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

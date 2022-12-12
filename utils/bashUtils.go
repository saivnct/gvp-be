package utils

import (
	"os/exec"
)

func BashExecute(path string, args []string) (string, error) {
	var finalArgs = append([]string{path}, args...)
	cmd := exec.Command("/bin/sh", finalArgs...)

	stdout, err := cmd.Output()
	if err != nil {
		return "", err
	}
	output := string(stdout)
	return output, nil
}

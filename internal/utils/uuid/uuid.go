package uuid

import (
	"os/exec"
	"strings"
)

func NewUuid() string { // Only works on Linux! TODO: add support for windows (or don't haha)
	uuid, err := exec.Command("uuidgen").Output()

	if err != nil {
		return ""
	}

	str := string(uuid)
	return strings.Replace(str, "\n", "", -1)
}

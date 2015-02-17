package service

import (
	"os"
	"path/filepath"
	"strings"
)

func getCommand(service Service) string {
	cmdName := service.Command[0]
	if !strings.HasPrefix("/", cmdName) {
		if _, err := os.Stat(filepath.Join(service.Directory, cmdName)); err == nil {
			cmdName = filepath.Join(service.Directory, cmdName)
		} else if _, err = os.Stat(filepath.Join(service.Directory, cmdName+".exe")); err == nil {
			cmdName = filepath.Join(service.Directory, cmdName+".exe")
		}
	}
	return cmdName
}

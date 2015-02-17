package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type (
	SystemDServiceManager struct {
		unitFilePath string
	}
)

func NewSystemDServiceManager(unitFilePath string) *SystemDServiceManager {
	return &SystemDServiceManager{unitFilePath}
}

func (sdsm *SystemDServiceManager) Install(service Service) error {
	name := service.Name
	dstPath := filepath.Join(sdsm.unitFilePath, name+".service")
	os.Remove(dstPath)

	estr := ""
	for k, v := range service.Environment {
		estr += "\"" + k + "=" + v + "\" "
	}

	cmdName := getCommand(service)
	os.Chmod(cmdName, 0777)

	err := ioutil.WriteFile(dstPath, []byte(`
[Unit]
Description=`+service.Name+`

[Service]
Environment=`+estr+`
ExecStart=`+cmdName+`

[Install]
WantedBy=multi-user.target
  `), 0644)
	if err != nil {
		return err
	}

	out, err := exec.Command("systemctl", "daemon-reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error reloading daemon: %v", string(out))
	}
	out, err = exec.Command("systemctl", "start", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error starting service: %v", string(out))
	}
	out, err = exec.Command("systemctl", "enable", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error enabling service: %v", string(out))
	}
	return nil
}

func (sdsm *SystemDServiceManager) Uninstall(name string) error {
	dstPath := filepath.Join(sdsm.unitFilePath, name+".service")
	exec.Command("systemctl", "disable", name).Run()
	exec.Command("systemctl", "stop", name).Run()
	os.Remove(dstPath)
	exec.Command("systemctl", "daemon-reload").Run()
	return nil
}

func (sdsm *SystemDServiceManager) Start() error {
	return nil
}

func (sdsm *SystemDServiceManager) Stop() error {
	return nil
}

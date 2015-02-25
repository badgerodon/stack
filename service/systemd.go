package service

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type (
	SystemDServiceManager struct {
		unitFilePath string
		userMode     bool
	}
)

func NewSystemDServiceManager(unitFilePath string, userMode bool) *SystemDServiceManager {
	return &SystemDServiceManager{unitFilePath, userMode}
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
WorkingDirectory=`+service.Directory+`
Restart=always

[Install]
WantedBy=multi-user.target
  `), 0644)
	if err != nil {
		return err
	}

	mode := "--system"
	if sdsm.userMode {
		mode = "--user"
	}

	out, err := exec.Command("systemctl", mode, "daemon-reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error reloading daemon: %v", string(out))
	}
	out, err = exec.Command("systemctl", mode, "start", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error starting service: %v", string(out))
	}
	out, err = exec.Command("systemctl", mode, "enable", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error enabling service: %v", string(out))
	}
	return nil
}

func (sdsm *SystemDServiceManager) List() ([]string, error) {
	mode := "--system"
	if sdsm.userMode {
		mode = "--user"
	}

	out, err := exec.Command("systemctl", mode, "list-units", "--full", "--no-pager", "badgerodon-stack*").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing services: %v", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	services := []string{}
	i := 0
	for scanner.Scan() {
		if i > 0 {
			fs := strings.Fields(scanner.Text())
			if len(fs) == 0 {
				break
			}
			n := fs[0]
			if strings.Contains(n, ".") {
				n = n[:strings.Index(n, ".")]
			}
			services = append(services, n)
		}
		i++
	}
	return services, nil
}

func (sdsm *SystemDServiceManager) Uninstall(name string) error {
	mode := "--system"
	if sdsm.userMode {
		mode = "--user"
	}

	dstPath := filepath.Join(sdsm.unitFilePath, name+".service")
	exec.Command("systemctl", mode, "disable", name).Run()
	exec.Command("systemctl", mode, "stop", name).Run()
	os.Remove(dstPath)
	exec.Command("systemctl", mode, "daemon-reload").Run()
	return nil
}

func (sdsm *SystemDServiceManager) Start() error {
	return nil
}

func (sdsm *SystemDServiceManager) Stop() error {
	return nil
}

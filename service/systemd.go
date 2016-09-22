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
	// A SystemDManager manages services in systemd
	SystemDManager struct {
		unitFilePath string
		userMode     bool
	}
)

// NewSystemDManager creates a new systemd service manager
func NewSystemDManager(unitFilePath string, userMode bool) *SystemDManager {
	return &SystemDManager{unitFilePath, userMode}
}

func (mgr *SystemDManager) mode() string {
	mode := "--system"
	if mgr.userMode {
		mode = "--user"
	}
	return mode
}

func (mgr *SystemDManager) String() string {
	return fmt.Sprintf("SystemDManager(unit-files=%s user-mode=%v)",
		mgr.unitFilePath, mgr.userMode)
}

// Install installs the service
func (mgr *SystemDManager) Install(service Service) error {
	name := service.Name
	dstPath := filepath.Join(mgr.unitFilePath, name+".service")
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
ExecStart=`+cmdName+` `+strings.Join(service.Command[1:], " ")+`
WorkingDirectory=`+service.Directory+`
Restart=always

[Install]
WantedBy=multi-user.target
  `), 0644)
	if err != nil {
		return err
	}
	out, err := exec.Command("systemctl", mgr.mode(), "daemon-reload").CombinedOutput()
	if err != nil {
		return fmt.Errorf("error reloading daemon: %v", string(out))
	}
	out, err = exec.Command("systemctl", mgr.mode(), "start", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error starting service: %v", string(out))
	}
	out, err = exec.Command("systemctl", mgr.mode(), "enable", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error enabling service: %v", string(out))
	}
	return nil
}

// List lists the installed services
func (mgr *SystemDManager) List() ([]string, error) {
	out, err := exec.Command("systemctl", mgr.mode(), "list-units", "--all", "--full", "--no-pager", "--no-legend", "stack-*").CombinedOutput()
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

// Uninstall uninstall a service
func (mgr *SystemDManager) Uninstall(name string) error {
	dstPath := filepath.Join(mgr.unitFilePath, name+".service")
	exec.Command("systemctl", mgr.mode(), "disable", name).Run()
	exec.Command("systemctl", mgr.mode(), "stop", name).Run()
	os.Remove(dstPath)
	exec.Command("systemctl", mgr.mode(), "daemon-reload").Run()
	return nil
}

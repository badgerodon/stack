package service

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type UpstartServiceManager struct {
}

func NewUpstartServiceManager() *UpstartServiceManager {
	return &UpstartServiceManager{}
}

func (usm *UpstartServiceManager) Install(service Service) error {
	cmdName := getCommand(service)
	os.Chmod(cmdName, 0777)

	estr := ""
	for k, v := range service.Environment {
		estr += "\"" + k + "=" + v + "\" "
	}
	src := `
description "` + service.Name + `"

start on (started networking)
respawn

chdir ` + service.Directory + `
exec env ` + estr + ` ` + cmdName + `
`

	err := ioutil.WriteFile("/etc/init/"+service.Name+".conf", []byte(src), 0644)
	if err != nil {
		return err
	}

	bs, err := exec.Command("initctl", "start", service.Name).CombinedOutput()
	if err != nil {
		if strings.Contains(string(bs), "already running") {
			exec.Command("initctl", "stop", service.Name).Run()
			time.Sleep(10 * time.Second)
			bs, err = exec.Command("initctl", "start", service.Name).CombinedOutput()
		}
	}
	if err != nil {
		return fmt.Errorf("failed to start service: %v", string(bs))
	}

	return nil
}

func (usm *UpstartServiceManager) Uninstall(name string) error {
	os.Remove("/etc/init/" + name + ".conf")
	bs, err := exec.Command("initctl", "stop", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service: %v", string(bs))
	}
	return fmt.Errorf("not implemented")
}

func (usm *UpstartServiceManager) List() ([]string, error) {
	out, err := exec.Command("initctl", "list").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error listing services: %v", string(out))
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	services := []string{}
	for scanner.Scan() {
		fs := strings.Fields(scanner.Text())
		if len(fs) < 2 {
			continue
		}
		if strings.HasPrefix(fs[0], "badgerodon-stack-") && strings.Contains(fs[1], "running") {
			services = append(services, fs[0])
		}
	}
	return services, nil
}

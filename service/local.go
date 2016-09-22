package service

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"sync"

	"github.com/badgerodon/stack/service/runner"
	"github.com/kardianos/osext"
)

type (
	// A LocalManager manages services with a custom service runner
	LocalManager struct {
		stateFile string
		client    *rpc.Client
		mu        sync.Mutex
	}
)

// NewLocalManager creates a new local manager
func NewLocalManager(stateFile string) *LocalManager {
	lsm := &LocalManager{
		stateFile: stateFile,
	}
	return lsm
}

func (lsm *LocalManager) String() string {
	return fmt.Sprintf("LocalManager(state-file=%s)", lsm.stateFile)
}

func (lsm *LocalManager) call(serviceMethod string, args interface{}, reply interface{}) error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	if lsm.client == nil {
		exe, err := osext.Executable()
		if err != nil {
			return err
		}

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		defer listener.Close()

		runner := exec.Command(exe, "service-runner", "--address", listener.Addr().String(), "--state-file", lsm.stateFile)
		runner.Stdout = os.Stdout
		runner.Stderr = os.Stderr
		go func() {
			runner.Run()
			lsm.mu.Lock()
			lsm.client = nil
			lsm.mu.Unlock()
		}()

		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		lsm.client = rpc.NewClient(conn)
	}

	return lsm.client.Call(serviceMethod, args, reply)
}

// Install installs the service
func (lsm *LocalManager) Install(service Service) error {
	req := runner.InstallRequest{
		Service: runner.Service{
			Name:        service.Name,
			Directory:   service.Directory,
			Command:     service.Command,
			Environment: service.Environment,
		},
	}
	var res runner.InstallResult
	err := lsm.call("Runner.Install", &req, &res)
	if err != nil {
		return err
	}
	return nil
}

// Uninstall uninstalls the service
func (lsm *LocalManager) Uninstall(name string) error {
	req := runner.UninstallRequest{
		Name: name,
	}
	var res runner.UninstallResult
	err := lsm.call("Runner.Uninstall", &req, &res)
	if err != nil {
		return err
	}
	return nil
}

// List lists the installed services
func (lsm *LocalManager) List() ([]string, error) {
	req := runner.ListRequest{}
	var res runner.ListResult
	err := lsm.call("Runner.List", &req, &res)
	if err != nil {
		return nil, err
	}
	return res.Names, nil
}

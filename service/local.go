package service

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type (
	LocalService struct {
		Service
		PID int
	}
	LocalServiceManager struct {
		stateFilePath string
		state         map[string]LocalService
		ticker        *time.Ticker
		mu            sync.Mutex
	}
)

var LocalServiceManagerTickerDuration = time.Second * 10

func NewLocalServiceManager(stateFilePath string) *LocalServiceManager {
	return &LocalServiceManager{
		stateFilePath: stateFilePath,
		state:         map[string]LocalService{},
		ticker:        time.NewTicker(LocalServiceManagerTickerDuration),
	}
}

func (lsm *LocalServiceManager) readState() error {
	f, err := os.Open(lsm.stateFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewDecoder(f).Decode(&lsm.state)
}

func (lsm *LocalServiceManager) saveState() error {
	f, err := os.Create(lsm.stateFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(lsm.state)
}

func (lsm *LocalServiceManager) run(service Service) (int, error) {
	cmdName := getCommand(service)
	os.Chmod(cmdName, 0777)

	cmd := exec.Command(cmdName, service.Command[1:]...)
	cmd.Dir = service.Directory
	// can't use a map here because order actually matters... at least in windows anyway
	cmd.Env = []string{}
	for _, e := range os.Environ() {
		k := e[:strings.IndexByte(e, '=')]
		if _, ok := service.Environment[k]; !ok {
			cmd.Env = append(cmd.Env, e)
		}
	}
	for k, v := range service.Environment {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("[LocalServiceManager]", service.Name, "failed to create stdout pipe", service.Name)
		return 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		log.Println("[LocalServiceManager]", service.Name, "failed to create stderr pipe", service.Name)
		return 0, err
	}

	go func() {
		s := bufio.NewScanner(stdout)
		for s.Scan() {
			log.Println("["+service.Name+"]", s.Text())
		}
	}()
	go func() {
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			log.Println("["+service.Name+"]", s.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Println("[LocalServiceManager]", service.Name, "failed to start:", err)
		return 0, err
	}

	go func() {
		err := cmd.Wait()
		log.Println("[LocalServiceManager]", service.Name, "exited:", err)
	}()

	return cmd.Process.Pid, nil
}

func (lsm *LocalServiceManager) checkProcesses() error {
	changed := false
	for name, ls := range lsm.state {
		_, err := os.FindProcess(ls.PID)
		if err != nil {
			log.Printf("[LocalServiceManager] %v is dead, restarting\n", name)
			pid, err := lsm.run(ls.Service)
			if err == nil {
				log.Printf("[LocalServiceManager] started %v: %v\n", name, pid)
				changed = true
				lsm.state[name] = LocalService{ls.Service, pid}
			} else {
				log.Printf("[LocalServiceManager] error starting %v: %v\n", name, err)
			}
		}
	}
	if changed {
		return lsm.saveState()
	}
	return nil
}

func (lsm *LocalServiceManager) Install(service Service) error {
	log.Printf("[LocalServiceManager] installing %v\n", service.Name)

	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	lsm.state[service.Name] = LocalService{service, 0}
	return lsm.checkProcesses()
}

func (lsm *LocalServiceManager) Uninstall(name string) error {
	log.Printf("[LocalServiceManager] uninstalling %v\n", name)

	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	if ls, ok := lsm.state[name]; ok {
		proc, err := os.FindProcess(ls.PID)
		if err == nil {
			proc.Kill()
		}
		delete(lsm.state, name)
	}

	return lsm.saveState()
}

func (lsm *LocalServiceManager) List() ([]string, error) {
	panic("not implemented")
}

func (lsm *LocalServiceManager) Start() error {
	lsm.mu.Lock()
	defer lsm.mu.Unlock()

	lsm.readState()
	err := lsm.saveState()
	if err != nil {
		return err
	}

	go func() {
		for _ = range lsm.ticker.C {
			lsm.mu.Lock()
			lsm.checkProcesses()
			lsm.mu.Unlock()
		}
	}()

	return nil
}

func (lsm *LocalServiceManager) Stop() error {
	lsm.ticker.Stop()
	return nil
}

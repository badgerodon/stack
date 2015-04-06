package runner

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type (
	Runner struct {
		addr      string
		stateFile string
		services  map[string]int
		mu        sync.Mutex
	}
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

func kill(pid int) {
	proc, err := os.FindProcess(pid)
	if err == nil {
		proc.Kill()
		proc.Release()
	}
}

func Run(addr string, stateFile string) error {
	r := &Runner{
		addr:      addr,
		stateFile: stateFile,
		services:  make(map[string]int),
	}

	for name, svc := range r.loadState() {
		pid, err := r.run(svc)
		if err == nil {
			r.services[name] = pid
		}
	}

	server := rpc.NewServer()
	server.Register(r)

	log.Println("[runner] connecting to", r.addr)
	conn, err := net.Dial("tcp", r.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	server.ServeConn(conn)

	// cleanup
	r.mu.Lock()
	for _, pid := range r.services {
		kill(pid)
	}
	r.services = make(map[string]int)
	r.mu.Unlock()

	return nil
}

type (
	ListRequest struct{}
	ListResult  struct {
		Names []string
	}
)

func (r *Runner) List(req *ListRequest, res *ListResult) error {
	res.Names = []string{}
	r.mu.Lock()
	for name, _ := range r.services {
		res.Names = append(res.Names, name)
	}
	r.mu.Unlock()
	sort.Strings(res.Names)
	return nil
}

type (
	InstallRequest struct {
		Service
	}
	InstallResult struct {
	}
	Service struct {
		Name        string
		Directory   string
		Command     []string
		Environment map[string]string
	}
)

func (r *Runner) loadState() map[string]Service {
	services := map[string]Service{}

	f, err := os.Open(r.stateFile)
	if err != nil {
		return services
	}
	defer f.Close()

	json.NewDecoder(f).Decode(&services)

	return services
}

func (r *Runner) saveState(services map[string]Service) {
	f, err := os.Create(r.stateFile)
	if err != nil {
		return
	}
	defer f.Close()

	json.NewEncoder(f).Encode(&services)
}

func (r *Runner) run(service Service) (int, error) {
	cmdName := getCommand(service)
	cmd := exec.Command(cmdName, service.Command[1:]...)
	cmd.Dir = service.Directory
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
		log.Println("[runner]", service.Name, "failed to create stdout pipe", service.Name)
		return 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		log.Println("[runner]", service.Name, "failed to create stderr pipe", service.Name)
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
		log.Println("[runner]", service.Name, "failed to start:", err)
		return 0, err
	}

	pid := cmd.Process.Pid

	go func() {
		err := cmd.Wait()
		log.Println("[runner]", service.Name, "exited:", err)
		time.Sleep(time.Second * 10)
		r.mu.Lock()
		pidNow, ok := r.services[service.Name]
		r.mu.Unlock()
		if ok && pidNow == pid {
			log.Println("[runner] restarting", service.Name)
			r.run(service)
		}
	}()

	log.Println("[runner] started", service.Name, "pid=", cmd.Process.Pid)

	return pid, nil
}

func (r *Runner) Install(req *InstallRequest, res *InstallResult) error {
	// if it's already running, kill it
	r.mu.Lock()
	pid, ok := r.services[req.Name]
	if ok {
		delete(r.services, req.Name)
		kill(pid)
	}
	r.mu.Unlock()

	pid, err := r.run(req.Service)
	if err != nil {
		return err
	}

	r.mu.Lock()
	_, ok = r.services[req.Name]
	if ok {
		kill(pid)
	} else {
		r.services[req.Name] = pid
		services := r.loadState()
		services[req.Name] = req.Service
		r.saveState(services)
	}
	r.mu.Unlock()

	return nil
}

type (
	UninstallRequest struct {
		Name string
	}
	UninstallResult struct{}
)

func (r *Runner) Uninstall(req *UninstallRequest, res *UninstallResult) error {
	// if it's already running, kill it
	r.mu.Lock()
	pid, ok := r.services[req.Name]
	if ok {
		delete(r.services, req.Name)
		kill(pid)
	}
	services := r.loadState()
	delete(r.services, req.Name)
	r.saveState(services)
	r.mu.Unlock()
	return nil
}

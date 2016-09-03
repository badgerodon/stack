package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/badgerodon/stack/service"
	"github.com/badgerodon/stack/storage"
	"github.com/minio/blake2b-simd"
)

var rootDir, tmpDir string
var serviceManager service.ServiceManager

func isUpstart() bool {
	bs, err := exec.Command("/sbin/init", "--version").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(bs), "upstart")
}

func isSystemD() bool {
	bs, err := exec.Command("systemctl").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(bs), ".mount")
}

func init() {
	isRoot := os.Getuid() == 0
	switch runtime.GOOS {
	case "darwin":
		if isRoot {
			rootDir = "/opt/badgerodon-stack"
		} else {
			rootDir = filepath.Join(os.Getenv("HOME"), "badgerodon-stack")
		}
		serviceManager = service.NewLocalServiceManager(filepath.Join(rootDir, "services.state"))
	case "linux":
		if isRoot {
			rootDir = "/opt/badgerodon-stack"
			if isUpstart() {
				serviceManager = service.NewUpstartServiceManager()
			} else if isSystemD() {
				if _, err := os.Stat("/usr/lib/systemd/system"); err == nil {
					serviceManager = service.NewSystemDServiceManager("/usr/lib/systemd/system/", false)
				} else {
					serviceManager = service.NewSystemDServiceManager("/etc/systemd/system/", false)
				}
			} else {
				serviceManager = service.NewLocalServiceManager(filepath.Join(rootDir, "services.state"))
			}
		} else {
			rootDir = filepath.Join(os.Getenv("HOME"), "badgerodon-stack")
			serviceManager = service.NewLocalServiceManager(filepath.Join(rootDir, "services.state"))
		}
	case "windows":
		// TODO: check for access
		rootDir = "C:\\ProgramData\\badgerodon-stack"
		serviceManager = service.NewLocalServiceManager(filepath.Join(rootDir, "services.state"))
	default:
		panic("unsupported operating system")
	}

	os.MkdirAll(rootDir, 0755)
	for _, d := range []string{"applications", "downloads", "run", "tmp"} {
		os.MkdirAll(filepath.Join(rootDir, d), 0755)
	}
	tmpDir = filepath.Join(rootDir, "tmp")
}

type (
	// StackState is the local state of the badgerodon stack
	StackState struct {
		Applications []Application `yaml:"applications"`
		Downloads    map[string]string
	}

	Config struct {
		Applications []Application `yaml:"applications"`
	}
	Application struct {
		Name    string             `yaml:"name"`
		Source  storage.Location   `yaml:"source"`
		Links   map[string]string  `yaml:"links,omitempty"`
		Files   map[string]string  `yaml:"files,omitempty"`
		Service ApplicationService `yaml:"service,omitempty"`
	}
	ApplicationService struct {
		Command     []string          `yaml:"command,omitempty"`
		Environment map[string]string `yaml:"environment,omitempty"`
	}
)

// UnmarshalYAML unmarshals a yaml structure
func (as *ApplicationService) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var t1 struct {
		Command     []string          `yaml:"command,omitempty"`
		Environment map[string]string `yaml:"environment,omitempty"`
	}
	err := unmarshal(&t1)
	if err == nil {
		as.Command = t1.Command
		as.Environment = t1.Environment
		return nil
	}
	var t2 struct {
		Command     string            `yaml:"command,omitempty"`
		Environment map[string]string `yaml:"environment,omitempty"`
	}
	err = unmarshal(&t2)
	if err == nil {
		as.Command = strings.Fields(t2.Command)
		as.Environment = t2.Environment
		return nil
	}
	return err
}

func (a Application) ApplicationPath() string {
	return filepath.Join(rootDir, "applications", a.Name)
}

func (a Application) DownloadPath() string {
	return filepath.Join(rootDir, "downloads", a.Name+a.Source.Ext())
}

// Hash is a hash of the application data
func (a Application) Hash() string {
	bs, _ := json.Marshal(a)
	return fmt.Sprintf("%X", blake2b.Sum512(bs))
}

// SourceHash is a hash of the source location
func (a Application) SourceHash() string {
	bs, _ := json.Marshal(a.Source)
	return fmt.Sprintf("%X", blake2b.Sum512(bs))
}

func (a Application) ServiceName() string {
	return "badgerodon-stack-" + a.Name
}

func ReadStackState() *StackState {
	state := &StackState{}
	bs, err := ioutil.ReadFile(filepath.Join(rootDir, "state.json"))
	if err == nil {
		err = json.Unmarshal(bs, state)
		if err != nil {
			log.Println("[ReadStackState] error unmarshaling:", err)
		}
	}

	if state.Applications == nil {
		state.Applications = make([]Application, 0)
	}
	if state.Downloads == nil {
		state.Downloads = make(map[string]string)
	}

	Validate(state)

	return state
}

func SaveStackState(state *StackState) {
	out, err := json.Marshal(state)
	if err != nil {
		log.Println("[SaveStackState] error marshaling:", err)
		return
	}
	ioutil.WriteFile(filepath.Join(rootDir, "state.json"), out, 0755)
	log.Println("[SaveStackState] saved state:", string(out))
}

func ParseConfig(rdr io.Reader) (*Config, error) {
	bs, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, yaml.Unmarshal(bs, &cfg)
}

func Validate(state *StackState) {
	// applications
	root := filepath.Join(rootDir, "applications")
	existingApplications := map[string]struct{}{}
	filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if p == root {
			return nil
		}
		if !fi.IsDir() {
			log.Println("[config] removing file in applications:", p)
			os.Remove(p)
			return nil
		}
		existingApplications[p] = struct{}{}
		return filepath.SkipDir
	})
	// services
	existingServices := map[string]struct{}{}
	services, err := serviceManager.List()
	if err != nil {
		log.Println("[config] error listing services:", err)
	} else {
		for _, svc := range services {
			existingServices[svc] = struct{}{}
		}
	}

	var applicationsToRemove []string
	var servicesToRemove []string

	for i := 0; i < len(state.Applications); i++ {
		a := state.Applications[i]
		_, foundApplication := existingApplications[a.ApplicationPath()]
		_, foundService := existingServices[a.ServiceName()]
		if !(foundApplication && foundService && len(a.Service.Command) > 0) {
			log.Println("[config] removing invalid application", a.Name)
			copy(state.Applications[i:], state.Applications[i+1:])
			state.Applications = state.Applications[:len(state.Applications)-1]
			i--

			if foundApplication {
				applicationsToRemove = append(applicationsToRemove, a.ApplicationPath())
			}
			if foundService {
				servicesToRemove = append(servicesToRemove, a.ServiceName())
			}
		}
	}

	for _, a := range applicationsToRemove {
		log.Println("[config] removing untracked application", a)
		os.RemoveAll(a)
	}
	for _, s := range servicesToRemove {
		log.Println("[config] removing untracked service", s)
		serviceManager.Uninstall(s)
	}
}

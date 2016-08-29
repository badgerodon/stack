package main

import (
	"encoding/json"
	"hash/fnv"
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
				serviceManager = service.NewSystemDServiceManager("/usr/lib/systemd/system/", false)
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

func (a Application) Hash() uint64 {
	h := fnv.New64a()
	json.NewEncoder(h).Encode(a)
	return h.Sum64()
}

func (a Application) SourceHash() uint64 {
	h := fnv.New64a()
	json.NewEncoder(h).Encode(a.Source)
	return h.Sum64()
}

func (a Application) ServiceName() string {
	return "badgerodon-stack-" + a.Name
}

func ReadConfig() *Config {
	cfg := &Config{}
	bs, err := ioutil.ReadFile(filepath.Join(rootDir, "config.yaml"))
	if err == nil {
		err = yaml.Unmarshal(bs, cfg)
		if err != nil {
			log.Println("[ReadConfig] error unmarshaling:", err)
		}
	}

	if cfg.Applications == nil {
		cfg.Applications = make([]Application, 0)
	}

	Validate(cfg)

	return cfg
}

func SaveConfig(cfg *Config) {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		log.Println("[SaveConfig] error marshaling:", err)
		return
	}
	ioutil.WriteFile(filepath.Join(rootDir, "config.yaml"), out, 0755)
}

func ParseConfig(rdr io.Reader) (*Config, error) {
	bs, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, yaml.Unmarshal(bs, &cfg)
}

func Validate(cfg *Config) {
	// downloads
	root := filepath.Join(rootDir, "downloads")
	existingDownloads := map[string]struct{}{}
	filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if p == root {
			return nil
		}
		if fi.IsDir() {
			log.Println("[config] removing folder in downloads:", p)
			os.RemoveAll(p)
			return filepath.SkipDir
		}
		existingDownloads[p] = struct{}{}
		return nil
	})
	// applications
	root = filepath.Join(rootDir, "applications")
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

	for i := 0; i < len(cfg.Applications); i++ {
		a := cfg.Applications[i]
		_, foundDownload := existingDownloads[a.DownloadPath()]
		_, foundApplication := existingApplications[a.ApplicationPath()]
		_, foundService := existingServices[a.ServiceName()]
		if foundDownload && foundApplication && (foundService || len(a.Service.Command) == 0) {
			delete(existingDownloads, a.DownloadPath())
			delete(existingApplications, a.ApplicationPath())
			delete(existingServices, a.ServiceName())
		} else {
			log.Println("[config] removing invalid application", a.Name)
			copy(cfg.Applications[i:], cfg.Applications[i+1:])
			cfg.Applications = cfg.Applications[:len(cfg.Applications)-1]
			i--
		}
	}
	for a := range existingApplications {
		log.Println("[config] removing untracked application", a)
		os.RemoveAll(a)
	}
	for p := range existingDownloads {
		log.Println("[config] removing untracked download", p)
		os.Remove(p)
	}
	for s := range existingServices {
		log.Println("[config] removing untracked service", s)
		serviceManager.Uninstall(s)
	}
}

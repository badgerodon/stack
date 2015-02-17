package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/badgerodon/stack/service"
)

func ProcessApplications(prevCfg, newCfg *Config) error {
	for _, pa := range prevCfg.Applications {
		found := false
		for _, na := range newCfg.Applications {
			if pa.Hash() == na.Hash() {
				found = true
				break
			}
		}
		if !found {
			log.Println("[applications] remove service", pa.ServiceName())
			err := serviceManager.Uninstall(pa.ServiceName())
			if err != nil {
				return err
			}

			log.Println("[applications] remove folder", pa.ApplicationPath())
			err = os.RemoveAll(pa.ApplicationPath())
			if err != nil {
				return err
			}
		}
	}
	for _, na := range newCfg.Applications {
		found := false
		for _, pa := range prevCfg.Applications {
			if na.Hash() == pa.Hash() {
				found = true
				break
			}
		}
		if found {
			log.Println("[applications] skipping", na.Name)
		} else {
			log.Println("[applications] extract folder", na.ApplicationPath())
			err := extract(na.ApplicationPath(), na.DownloadPath())
			if err != nil {
				return err
			}

			for name, target := range na.Links {
				fp := filepath.Join(na.ApplicationPath(), name)
				tp := filepath.Join(na.ApplicationPath(), target)
				log.Println("[applications] add link", fp)
				err := os.Link(tp, fp)
				if err != nil {
					return err
				}
			}
			for name, content := range na.Files {
				fp := filepath.Join(na.ApplicationPath(), name)
				log.Println("[applications] add file", fp)
				err := ioutil.WriteFile(fp, []byte(content), 0755)
				if err != nil {
					return err
				}
			}

			log.Println("[applications] install service", na.ServiceName())
			err = serviceManager.Install(service.Service{
				Name:        na.ServiceName(),
				Directory:   na.ApplicationPath(),
				Command:     na.Service.Command,
				Environment: na.Service.Environment,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

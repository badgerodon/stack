package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/badgerodon/stack/archive"
	"github.com/badgerodon/stack/service"
	"github.com/badgerodon/stack/storage"
)

func apply(src string) error {
	pl := NewPortLock(49001)
	pl.Lock()
	defer pl.Unlock()

	loc, err := storage.ParseLocation(src)
	if err != nil {
		return err
	}
	rc, err := storage.Get(loc)
	if err != nil {
		return err
	}
	defer rc.Close()

	cfg, err := ParseConfig(rc)
	if err != nil {
		return err
	}

	prevCfg := ReadConfig()

	err = applySources(prevCfg, cfg)
	if err != nil {
		return fmt.Errorf("error processing sources: %v", err)
	}

	err = applyApplications(prevCfg, cfg)
	if err != nil {
		return fmt.Errorf("error processing applications: %v", err)
	}

	SaveConfig(cfg)

	return nil
}

func applyApplications(prevCfg, newCfg *Config) error {
	for _, pa := range prevCfg.Applications {
		found := false
		for _, na := range newCfg.Applications {
			if pa.Hash() == na.Hash() {
				found = true
				break
			}
		}
		if !found {
			log.Println("[install] [application] remove service", pa.ServiceName())
			err := serviceManager.Uninstall(pa.ServiceName())
			if err != nil {
				return err
			}

			log.Println("[install] [application] remove folder", pa.ApplicationPath())
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
			log.Println("[install] [application] skip", na.Name)
		} else {
			log.Println("[install] [application] extract folder", na.ApplicationPath())
			err := archive.Extract(na.ApplicationPath(), na.DownloadPath())
			if err != nil {
				return fmt.Errorf("error extracting folder: %v", err)
			}

			for name, target := range na.Links {
				fp := filepath.Join(na.ApplicationPath(), name)
				tp := filepath.Join(na.ApplicationPath(), target)
				log.Println("[install] [application] add link", fp)
				err := os.Link(tp, fp)
				if err != nil {
					return fmt.Errorf("error creating link: %v", err)
				}
			}
			for name, content := range na.Files {
				fp := filepath.Join(na.ApplicationPath(), name)
				log.Println("[install] [application] add file", fp)
				err := ioutil.WriteFile(fp, []byte(content), 0755)
				if err != nil {
					return fmt.Errorf("error creating file: %v", err)
				}
			}

			if len(na.Service.Command) > 0 {
				log.Println("[install] [application] install service", na.ServiceName())
				err = serviceManager.Install(service.Service{
					Name:        na.ServiceName(),
					Directory:   na.ApplicationPath(),
					Command:     na.Service.Command,
					Environment: na.Service.Environment,
				})
				if err != nil {
					return fmt.Errorf("error installing service: %v", err)
				}
			}
		}
	}
	return nil
}

func applySources(prevCfg, newCfg *Config) error {
	// remove
	for _, pa := range prevCfg.Applications {
		found := false
		for _, na := range newCfg.Applications {
			if pa.SourceHash() == na.SourceHash() {
				found = true
				break
			}
		}
		if !found {
			log.Println("[install] [source] remove", pa.DownloadPath())
			os.Remove(pa.DownloadPath())
		}
	}
	// add
	for _, na := range newCfg.Applications {
		found := false
		for _, pa := range prevCfg.Applications {
			if na.SourceHash() == pa.SourceHash() {
				found = true
				break
			}
		}
		if found {
			log.Println("[install] [source] skip", na.DownloadPath())
		} else {
			log.Println("[install] [source] download", na.DownloadPath(), na.Source)
			rc, err := storage.Get(na.Source)
			if err != nil {
				return fmt.Errorf("error downloading: %v", err)
			}
			f, err := os.Create(na.DownloadPath())
			if err != nil {
				rc.Close()
				return fmt.Errorf("error creating download file: %v", err)
			}
			_, err = io.Copy(f, rc)
			rc.Close()
			f.Close()
			if err != nil {
				return fmt.Errorf("error downloading: %v", err)
			}
		}
	}
	return nil
}

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

	state := ReadStackState()

	err = applySources(state, cfg)
	if err != nil {
		return fmt.Errorf("error processing sources: %v", err)
	}

	err = applyApplications(state, cfg)
	if err != nil {
		return fmt.Errorf("error processing applications: %v", err)
	}

	return nil
}

func applyApplications(state *StackState, newCfg *Config) error {
	for _, pa := range state.Applications {
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
		for _, pa := range state.Applications {
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

			state.Applications = append(state.Applications, na)
			SaveStackState(state)
		}
	}
	return nil
}

func applySources(state *StackState, newCfg *Config) error {
	// remove
	for path, hash := range state.Downloads {
		found := false
		for _, app := range newCfg.Applications {
			if app.DownloadPath() == path && app.SourceHash() == hash {
				found = true
				break
			}
		}
		if !found {
			log.Println("[install] [source] remove", path)
			os.Remove(path)

			delete(state.Downloads, path)
			SaveStackState(state)
		}
	}
	// add
	for _, app := range newCfg.Applications {
		path := app.DownloadPath()
		hash := app.SourceHash()
		if _, ok := state.Downloads[path]; ok {
			// no need to check hash because we would have already removed the file
			// above
			continue
		}
		log.Println("[install] [source] download", path, app.Source)

		rc, err := storage.Get(app.Source)
		if err != nil {
			return fmt.Errorf("error downloading: %v", err)
		}
		//TODO: make this an atomic update
		f, err := os.Create(path)
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

		state.Downloads[path] = hash
		SaveStackState(state)
	}
	return nil
}

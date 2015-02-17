package sync

import (
	"fmt"
	"net/url"
	"runtime"
)

type (
	Source struct {
		underlying interface{}
		syncer     Syncer
	}
	Syncer interface {
		Watch() *Watcher
		Sync(dst string) error
		Ext() string
	}
)

func (src Source) MarshalYAML() (interface{}, error) {
	return src.underlying, nil
}

func (src *Source) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&src.underlying)
}

func Parse(obj interface{}) (Syncer, error) {
	switch t := obj.(type) {
	case string:
		u, err := url.Parse(t)
		if err != nil {
			return nil, err
		}
		switch u.Scheme {
		case "http":
			fallthrough
		case "https":
			return HTTPSource{t}, nil
		case "local":
			fallthrough
		case "file":
			p := t[len("file://"):]
			if runtime.GOOS == "windows" {
				p = p[1:]
			}
			return LocalSource{p}, nil
		case "mega":
			return NewMegaSourceFromString(t), nil
		default:
			return nil, fmt.Errorf("unknown source: %v", t)
		}
	default:
		return nil, fmt.Errorf("don't know how to parse objects of type %T", obj)
	}
}

func (src Source) Sync(dst string) error {
	s, err := Parse(src.underlying)
	if err != nil {
		return err
	}
	return s.Sync(dst)
}

func (src Source) Ext() string {
	s, err := Parse(src.underlying)
	if err != nil {
		return ""
	}
	return s.Ext()
}

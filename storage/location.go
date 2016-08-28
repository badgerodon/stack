package storage

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

type (
	Location map[string]string
)

func (loc Location) Ext() string {
	p := loc.Path()
	ext := path.Ext(p)
	if strings.HasSuffix(p, ".tar"+ext) {
		ext = ".tar" + ext
	}
	return ext
}

func (loc Location) Host() string {
	return loc["host"]
}

func (loc Location) Path() string {
	return loc["path"]
}

func (loc Location) Type() string {
	return loc["type"]
}

func (loc *Location) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var obj interface{}
	err := unmarshal(&obj)
	if err != nil {
		return err
	}
	*loc, err = ParseLocation(obj)
	return err
}

// ParseLocation parses a location from a variety of input formats
func ParseLocation(obj interface{}) (Location, error) {
	switch t := obj.(type) {
	case string:
		if strings.Contains(t, "://") {
			u, err := url.Parse(t)
			if err != nil {
				return nil, fmt.Errorf("invalid location: %v", err)
			}
			loc := Location{
				"scheme": u.Scheme,
				"type":   u.Scheme,
				"host":   u.Host,
				"path":   u.Path,
				"query":  u.Query().Encode(),
			}
			if u.User != nil {
				loc["user"] = u.User.Username()
				loc["password"], _ = u.User.Password()
			}
			return loc, nil
		}
		// no scheme, so treat this as a local file
		return Location{
			"type": "local",
			"path": t,
		}, nil
	case map[interface{}]interface{}:
		loc := Location{}
		for k, v := range t {
			loc[fmt.Sprint(k)] = fmt.Sprint(v)
		}
		return loc, nil
	case map[string]interface{}:
		loc := Location{}
		for k, v := range t {
			loc[k] = fmt.Sprint(v)
		}
		return loc, nil
	case map[string]string:
		return Location(t), nil
	}
	return nil, fmt.Errorf("invalid location: %T", obj)
}

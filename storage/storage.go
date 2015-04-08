package storage

import (
	"fmt"
	"io"
	"log"
	"os"
)

type (
	AuthProvider interface {
		Authenticate()
	}
	Provider interface {
		Delete(location Location) error
		Get(location Location) (io.ReadCloser, error)
		Put(location Location, rdr io.Reader) error
		List(location Location) ([]string, error)
		Version(location Location, previous string) (string, error)
	}
	Sizer interface {
		Size() (int64, error)
	}
)

var (
	authProviders = map[string]AuthProvider{}
	providers     = map[string]Provider{}
)

func RegisterAuth(scheme string, authProvider AuthProvider) {
	authProviders[scheme] = authProvider
}

func Register(scheme string, provider Provider) {
	providers[scheme] = provider
}

type (
	DeleteOnClose struct {
		*os.File
	}
)

func (doc DeleteOnClose) Close() error {
	doc.File.Close()
	return os.Remove(doc.File.Name())
}

func getSize(rdr io.Reader) (int64, error) {
	if szr, ok := rdr.(Sizer); ok {
		return szr.Size()
	}
	return 0, fmt.Errorf("could not find size implementation")
}

func GetProvider(loc Location) (Provider, error) {
	p, ok := providers[loc.Type()]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("unknown storage provider: %v", loc.Type())
}

func Authenticate(typ string) {
	p, ok := authProviders[typ]
	if ok {
		p.Authenticate()
	} else {
		log.Fatalln("no auth provider registered for", typ)
	}
}

func Delete(loc Location) error {
	p, err := GetProvider(loc)
	if err != nil {
		return err
	}
	return p.Delete(loc)
}

func Get(loc Location) (io.ReadCloser, error) {
	p, err := GetProvider(loc)
	if err != nil {
		return nil, err
	}
	return p.Get(loc)
}

func List(loc Location) ([]string, error) {
	p, err := GetProvider(loc)
	if err != nil {
		return nil, err
	}
	return p.List(loc)
}

func Put(loc Location, rdr io.Reader) error {
	p, err := GetProvider(loc)
	if err != nil {
		return err
	}
	return p.Put(loc, rdr)
}

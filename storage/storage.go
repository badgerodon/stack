package storage

import (
	"fmt"
	"io"
	"log"
	"os"
)

type (
	// A Getter can get files
	Getter interface {
		Get(location Location) (io.ReadCloser, error)
	}

	// A Lister can list files
	Lister interface {
		List(Location) ([]string, error)
	}

	// A Putter can put files
	Putter interface {
		Put(location Location, rdr io.Reader) error
	}

	Versioner interface {
		Version(location Location, previous string) (string, error)
	}

	AuthProvider interface {
		Authenticate()
	}
	Provider interface {
		Getter
		Putter
		Lister
		Versioner

		Delete(location Location) error
	}
	Sizer interface {
		Size() (int64, error)
	}
)

func init() {
	Register("bitbucket", BitBucket)
	Register("github", GitHub)
	Register("gs", googleStorage{})
}

var (
	authProviders = map[string]AuthProvider{}
	providers     = map[string]Provider{}

	getters    = map[string]Getter{}
	putters    = map[string]Putter{}
	listers    = map[string]Lister{}
	versioners = map[string]Versioner{}
)

func RegisterAuth(scheme string, authProvider AuthProvider) {
	authProviders[scheme] = authProvider
}

// Register registers a provider
func Register(scheme string, provider interface{}) {
	if g, ok := provider.(Getter); ok {
		getters[scheme] = g
	}
	if p, ok := provider.(Putter); ok {
		putters[scheme] = p
	}
	if l, ok := provider.(Lister); ok {
		listers[scheme] = l
	}
	if v, ok := provider.(Versioner); ok {
		versioners[scheme] = v
	}
	if p, ok := provider.(Provider); ok {
		providers[scheme] = p
	}
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

// Get returns an io.ReadCloser for the given location
func Get(loc Location) (io.ReadCloser, error) {
	g, ok := getters[loc.Type()]
	if !ok {
		return nil, fmt.Errorf("no getter associated with scheme: %v", loc.Type())
	}
	return g.Get(loc)
}

// List returns a list of filenames for the given location
func List(loc Location) ([]string, error) {
	l, ok := listers[loc.Type()]
	if !ok {
		return nil, fmt.Errorf("no lister associated with schem: %v", loc.Type())
	}
	return l.List(loc)
}

// Put writes an io.Reader to the given location
func Put(loc Location, rdr io.Reader) error {
	p, ok := putters[loc.Type()]
	if !ok {
		return fmt.Errorf("no putter associated with scheme: %v", loc.Type())
	}
	return p.Put(loc, rdr)
}

// Version returns the version of the file at the given location
func Version(loc Location, previous string) (string, error) {
	v, ok := versioners[loc.Type()]
	if !ok {
		return "", fmt.Errorf("no versioner associated with scheme: %v", loc.Type())
	}
	return v.Version(loc, previous)
}

package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/t3rm1n4l/go-mega"
)

type (
	MegaProvider struct {
		clients chan megaClient
	}
	megaref struct {
		email, password, path string
	}
	megaClient struct {
		*mega.Mega
		provider        *MegaProvider
		email, password string
		expires         time.Time
	}
)

func (mc megaClient) Close() {
	select {
	case mc.provider.clients <- mc:
	default:
	}
}

var (
	Mega = &MegaProvider{
		clients: make(chan megaClient, 10),
	}
	MegaFileNotFound = fmt.Errorf("mega.co.nz file not found")
)

func init() {
	go func() {
		for range time.NewTicker(time.Minute).C {
			clients := make([]megaClient, 0, 10)
			for {
				select {
				case client := <-Mega.clients:
					if client.expires.After(time.Now()) {
						clients = append(clients, client)
					}
					continue
				default:
				}
				break
			}
			for _, client := range clients {
				select {
				case Mega.clients <- client:
				default:
				}
			}
		}
	}()

	Register("mega", Mega)
}

func (mp *MegaProvider) parse(loc Location) megaref {
	ref := megaref{}
	// email
	if ref.email == "" {
		ref.email = loc["username"]
	}
	if ref.email == "" {
		ref.email = loc["email"]
	}
	if ref.email == "" {
		ref.email = os.Getenv("MEGA_EMAIL")
	}
	// password
	if ref.password == "" {
		ref.password = loc["password"]
	}
	if ref.password == "" {
		ref.password = os.Getenv("MEGA_PASSWORD")
	}
	// path
	if ref.path == "" {
		ref.path = loc["path"]
	}
	for strings.HasPrefix(ref.path, "/") {
		ref.path = ref.path[1:]
	}
	for strings.HasSuffix(ref.path, "/") {
		ref.path = ref.path[:len(ref.path)-1]
	}

	if loc["host"] != "mega.co.nz" {
		if ref.path == "" {
			ref.path = loc["host"]
		} else {
			ref.path = loc["host"] + "/" + ref.path
		}
	}

	return ref
}

func (mp *MegaProvider) getClient(ref megaref) (megaClient, error) {
	toUse := megaClient{
		provider: mp,
		email:    ref.email,
		password: ref.password,
		expires:  time.Now().Add(time.Minute * 10),
	}
	clients := make([]megaClient, 0, 10)
	for {
		select {
		case client := <-mp.clients:
			if client.email == ref.email && client.password == ref.password {
				toUse = client
			} else {
				clients = append(clients, client)
				continue
			}
		default:
		}
		break
	}
	for _, client := range clients {
		select {
		case mp.clients <- client:
		default:
		}
	}

	if toUse.Mega != nil {
		return toUse, nil
	}

	toUse.Mega = mega.New()

	err := toUse.Login(ref.email, ref.password)
	if err != nil {
		return toUse, fmt.Errorf("mega.co.nz login failure: %v, email=%s", err, ref.email)
	}

	return toUse, nil
}

func (mp *MegaProvider) getNode(client *mega.Mega, path string, create bool) (*mega.Node, error) {
	parts := strings.Split(path, "/")
	node := client.FS.GetRoot()
	for _, part := range parts {
		children, err := client.FS.GetChildren(node)
		if err != nil {
			return nil, fmt.Errorf("failed to get node children: %v", err)
		}
		found := false
		for _, child := range children {
			if part == child.GetName() {
				node = child
				found = true
			}
		}
		if !found {
			if create {
				node, err = client.CreateDir(part, node)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, MegaFileNotFound
			}
		}
	}
	return node, nil
}

func (mp *MegaProvider) Delete(loc Location) error {
	ref := mp.parse(loc)
	client, err := mp.getClient(ref)
	if err != nil {
		return err
	}
	defer client.Close()

	node, err := mp.getNode(client.Mega, ref.path, false)
	if err != nil {
		return err
	}

	return client.Delete(node, true)
}

func (mp *MegaProvider) Get(loc Location) (io.ReadCloser, error) {
	ref := mp.parse(loc)
	client, err := mp.getClient(ref)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	node, err := mp.getNode(client.Mega, ref.path, false)
	if err != nil {
		return nil, err
	}

	tmpName := filepath.Join(os.TempDir(), uuid.New())

	err = client.DownloadFile(node, tmpName, nil)
	if err != nil {
		os.Remove(tmpName)
		return nil, fmt.Errorf("failed to download file: %v", err)
	}

	f, err := os.Open(tmpName)
	if err != nil {
		return nil, fmt.Errorf("failed to open temporary file: %v", err)
	}
	return DeleteOnClose{f}, nil
}

func (mp *MegaProvider) Put(loc Location, rdr io.Reader) error {
	ref := mp.parse(loc)
	client, err := mp.getClient(ref)
	if err != nil {
		return err
	}
	defer client.Close()

	name := ref.path
	if strings.Contains(name, "/") {
		name = name[strings.LastIndex(name, "/")+1:]
	}

	tmpName := filepath.Join(os.TempDir(), uuid.New())
	defer os.Remove(tmpName)
	tmp, err := os.Create(tmpName)
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	_, err = io.Copy(tmp, rdr)
	tmp.Close()
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}

	mp.Delete(loc)

	parent, err := mp.getNode(client.Mega, path.Dir(ref.path), true)
	if err != nil {
		return err
	}

	node, err := client.UploadFile(tmpName, parent, name, nil)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	err = client.Rename(node, name)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

func (mp *MegaProvider) List(loc Location) ([]string, error) {
	ref := mp.parse(loc)
	client, err := mp.getClient(ref)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	node, err := mp.getNode(client.Mega, ref.path, false)
	if err == MegaFileNotFound {
		return []string{}, nil
	} else if err != nil {
		return nil, err
	}

	children, err := client.FS.GetChildren(node)
	if err != nil {
		return nil, fmt.Errorf("error getting node children: %v", err)
	}
	names := []string{}
	for _, child := range children {
		names = append(names, child.GetName())
	}
	sort.Strings(names)
	return names, nil
}

func (mp *MegaProvider) Version(loc Location, previous string) (string, error) {
	ref := mp.parse(loc)
	log.Println(ref)
	client, err := mp.getClient(ref)
	if err != nil {
		return "", err
	}
	defer client.Close()

	node, err := mp.getNode(client.Mega, ref.path, false)
	if err != nil {
		return "", err
	}

	return node.GetHash(), nil
}

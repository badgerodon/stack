package storage

import (
	"fmt"
	"io"
	"net/url"
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

var Mega = &MegaProvider{
	clients: make(chan megaClient, 10),
}

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
}

func (mp *MegaProvider) parse(rawurl string) megaref {
	ref := megaref{
		email:    os.Getenv("MEGA_EMAIL"),
		password: os.Getenv("MEGA_PASSWORD"),
	}
	u, err := url.Parse(rawurl)
	if err == nil {
		if u.Host != "mega.co.nz" {
			ref.path = "/" + u.Host + u.Path
		} else {
			ref.path = u.Path
		}
		if u.User != nil {
			ref.email = u.User.Username()
			pw, ok := u.User.Password()
			if ok {
				ref.password = pw
			}
		}
	} else {
		ref.path = rawurl
	}
	if strings.HasPrefix(ref.path, "/") {
		ref.path = ref.path[1:]
	}
	if strings.HasSuffix(ref.path, "/") {
		ref.path = ref.path[:len(ref.path)-1]
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
		return toUse, fmt.Errorf("invalid username or password: %v", err)
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
				return nil, fmt.Errorf("file not found")
			}
		}
	}
	return node, nil
}

func (mp *MegaProvider) Delete(rawurl string) error {
	ref := mp.parse(rawurl)
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

func (mp *MegaProvider) Get(rawurl string) (io.ReadCloser, error) {
	ref := mp.parse(rawurl)
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

func (mp *MegaProvider) Put(rawurl string, rdr io.Reader) error {
	ref := mp.parse(rawurl)
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

	mp.Delete(rawurl)

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

func (mp *MegaProvider) List(rawurl string) ([]string, error) {
	ref := mp.parse(rawurl)
	client, err := mp.getClient(ref)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	node, err := mp.getNode(client.Mega, ref.path, false)
	if err != nil {
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

func (mp *MegaProvider) Version(rawurl string) (string, error) {
	ref := mp.parse(rawurl)
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

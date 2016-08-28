package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
)

// (gdocs|gdrive)://user[:password]@other.host/some_dir

type (
	GoogleDriveProvider byte
	googleDriveRef      struct {
		clientID     string
		clientSecret string
		token        oauth2.Token
		path         []string
	}
)

const (
	DefaultGoogleDriveClientID     = "304359942533-ra5badnhb5f1umi5vj4p5oohfhdiq8v8.apps.googleusercontent.com"
	DefaultGoogleDriveClientSecret = "2ORaxB_WysnMlfeYW5yZsBgH"
	GoogleDrive                    = GoogleDriveProvider(1)
)

var zeroTime time.Time

func init() {
	Register("gdocs", GoogleDrive)
	Register("gdrive", GoogleDrive)
	RegisterAuth("gdocs", GoogleDrive)
	RegisterAuth("gdrive", GoogleDrive)
}

func (gdp GoogleDriveProvider) parse(loc Location) googleDriveRef {
	ref := googleDriveRef{}

	ref.clientID = os.Getenv("GOOGLE_DRIVE_CLIENT_ID")
	ref.clientSecret = os.Getenv("GOOGLE_DRIVE_CLIENT_ID")
	ref.token.AccessToken = os.Getenv("GOOGLE_DRIVE_ACCESS_TOKEN")
	ref.token.TokenType = os.Getenv("GOOGLE_DRIVE_TOKEN_TYPE")
	ref.token.Expiry, _ = time.Parse(os.Getenv("GOOGLE_DRIVE_EXPIRY"), time.RFC3339)
	ref.token.RefreshToken = os.Getenv("GOOGLE_DRIVE_REFRESH_TOKEN")

	if ref.clientID == "" {
		ref.clientID = DefaultGoogleDriveClientID
	}

	if ref.clientSecret == "" {
		ref.clientSecret = DefaultGoogleDriveClientSecret
	}

	// support a json block for credentials
	if ref.token.RefreshToken == "" {
		json.Unmarshal([]byte(os.Getenv("GOOGLE_DRIVE_CREDENTIALS")), &ref.token)
	}

	// support a json file for credentials
	for _, nm := range []string{"GOOGLE_DRIVE_CREDENTIAL_FILE", "GOOGLE_DRIVE_CREDENTIALS_FILE"} {
		if ref.token.RefreshToken == "" && os.Getenv(nm) != "" {
			f, err := os.Open(os.Getenv(nm))
			if err == nil {
				defer f.Close()

				err = json.NewDecoder(f).Decode(&ref.token)
			}
		}
	}

	if ref.token.AccessToken == "" {
		ref.token.AccessToken = loc["access_token"]
	}
	if ref.token.RefreshToken == "" {
		ref.token.RefreshToken = loc["refresh_token"]
	}
	if ref.token.Expiry == zeroTime {
		ref.token.Expiry, _ = time.Parse(loc["expiry"], time.RFC3339)
	}
	if ref.token.TokenType == "" {
		ref.token.TokenType = loc["token_type"]
	}

	ref.path = []string{}
	if loc["host"] != "" {
		ref.path = append(ref.path, loc["host"])
	}
	for _, p := range strings.Split(loc["path"], "/") {
		if p != "" {
			ref.path = append(ref.path, p)
		}
	}

	return ref
}

func (gdp GoogleDriveProvider) allFiles(service *drive.Service, query string) ([]*drive.File, error) {
	fs := []*drive.File{}
	token := ""
	for {
		q := service.Files.List().Q(query)
		if token != "" {
			q = q.PageToken(token)
		}
		res, err := q.Do()
		if err != nil {
			return fs, err
		}
		fs = append(fs, res.Items...)
		token = res.NextPageToken
		if token == "" {
			break
		}
	}
	return fs, nil
}

func (gdp GoogleDriveProvider) newService(ref googleDriveRef) (*http.Client, *drive.Service, error) {
	config := &oauth2.Config{
		ClientID:     ref.clientID,
		ClientSecret: ref.clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveScope},
	}
	bg := context.Background()
	client := config.Client(bg, &ref.token)
	service, err := drive.New(client)
	return client, service, err
}

func (gdp GoogleDriveProvider) getFolderID(service *drive.Service, path []string) (string, error) {
	folderID := "root"
	for _, name := range path {
		fileList, err := gdp.allFiles(service, "'"+folderID+"' in parents and trashed = false and title = '"+strings.Replace(name, "'", "\\'", -1)+"' and mimeType = 'application/vnd.google-apps.folder'")
		if err != nil {
			return "", err
		}
		if len(fileList) < 1 {
			return "", nil
		}
		folderID = fileList[0].Id
	}
	return folderID, nil
}

func (gdp GoogleDriveProvider) getOrCreateFolderID(service *drive.Service, path []string) (string, error) {
	folderID := "root"
	for _, name := range path {
		fileList, err := gdp.allFiles(service, "'"+folderID+"' in parents and trashed = false and title = '"+strings.Replace(name, "'", "\\'", -1)+"' and mimeType = 'application/vnd.google-apps.folder'")
		if err != nil {
			return "", err
		}
		if len(fileList) < 1 {
			f, err := service.Files.Insert(&drive.File{
				Title:    name,
				Parents:  []*drive.ParentReference{{Id: folderID}},
				MimeType: "application/vnd.google-apps.folder",
			}).Do()
			if err != nil {
				return "", err
			}
			folderID = f.Id
		} else {
			folderID = fileList[0].Id
		}
	}
	return folderID, nil
}

func (gdp GoogleDriveProvider) getFile(service *drive.Service, path []string) (*drive.File, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("file not found")
	}
	folder := path[:len(path)-1]
	folderID, err := gdp.getFolderID(service, folder)
	if err != nil {
		return nil, err
	}
	fs, err := gdp.allFiles(service, "'"+folderID+"' in parents and trashed = false and title = '"+strings.Replace(path[len(path)-1], "'", "\\'", -1)+"'")
	if err != nil {
		return nil, err
	}
	if len(fs) < 1 {
		return nil, fmt.Errorf("file not found")
	}
	return fs[0], nil
}

func (gdp GoogleDriveProvider) Authenticate() {
	clientID := DefaultGoogleDriveClientID
	fmt.Print("Enter a Client ID [default: " + clientID + "]: ")
	fmt.Scanln(&clientID)

	clientSecret := DefaultGoogleDriveClientSecret
	fmt.Print("Enter a Client Secret [default: " + clientSecret + "]: ")
	fmt.Scanln(&clientSecret)

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{drive.DriveScope},
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
	}
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	url := config.AuthCodeURL(randState, oauth2.AccessTypeOffline)
	fmt.Println("Visit this URL:")
	fmt.Println("")
	fmt.Println(url)
	fmt.Println("")

	code := ""
	fmt.Print("Enter the code: ")
	fmt.Scanln(&code)

	if code == "" {
		log.Fatalln("expected code")
		return
	}

	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalln("error exchanging token:", err)
	}

	bs, _ := json.MarshalIndent(token, "", "  ")
	log.Println(string(bs))
}

func (gdp GoogleDriveProvider) Delete(loc Location) error {
	ref := gdp.parse(loc)
	_, service, err := gdp.newService(ref)
	if err != nil {
		return err
	}

	file, err := gdp.getFile(service, ref.path)
	if err != nil {
		return err
	}

	return service.Files.Delete(file.Id).Do()
}

func (gdp GoogleDriveProvider) Get(loc Location) (io.ReadCloser, error) {
	ref := gdp.parse(loc)
	client, service, err := gdp.newService(ref)
	if err != nil {
		return nil, err
	}

	file, err := gdp.getFile(service, ref.path)
	if err != nil {
		return nil, err
	}

	res, err := client.Get(file.DownloadUrl)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("error retrieving file: %d", res.StatusCode)
	}

	return res.Body, nil
}

func (gdp GoogleDriveProvider) List(loc Location) ([]string, error) {
	ref := gdp.parse(loc)
	_, service, err := gdp.newService(ref)
	if err != nil {
		return nil, err
	}

	folderID, err := gdp.getFolderID(service, ref.path)
	if err != nil {
		return nil, err
	}

	fileList, err := gdp.allFiles(service, "'"+folderID+"' in parents and trashed = false")
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, item := range fileList {
		names = append(names, item.Title)
	}
	return names, nil
}

func (gdp GoogleDriveProvider) Put(loc Location, rdr io.Reader) error {
	ref := gdp.parse(loc)
	_, service, err := gdp.newService(ref)
	if err != nil {
		return err
	}
	if len(ref.path) == 0 {
		return fmt.Errorf("expected file path")
	}
	folderID, err := gdp.getOrCreateFolderID(service, ref.path[:len(ref.path)-1])
	if err != nil {
		return err
	}
	_, err = service.Files.Insert(&drive.File{
		Title:   ref.path[len(ref.path)-1],
		Parents: []*drive.ParentReference{{Id: folderID}},
	}).Media(rdr).Do()
	return err
}

func (gdp GoogleDriveProvider) Version(loc Location, previous string) (string, error) {
	ref := gdp.parse(loc)
	_, service, err := gdp.newService(ref)
	if err != nil {
		return "", err
	}
	file, err := gdp.getFile(service, ref.path)
	if err != nil {
		return "", err
	}
	return file.Etag, nil
}

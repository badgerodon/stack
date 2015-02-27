package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type (
	CopyCloudStorageProvider struct{}
	copyref                  struct {
		path     string
		name     string
		user     string
		password string
	}
)

var CopyCloudStorage = &CopyCloudStorageProvider{}

func init() {
	Register("copy", CopyCloudStorage)
}

func (ccsp *CopyCloudStorageProvider) ref(loc Location) copyref {
	ref := copyref{}
	ref.path = loc["path"]
	if !strings.HasPrefix(ref.path, "/") {
		ref.path = "/" + ref.path
	}
	if loc["host"] != "copy.api.com" {
		ref.path = loc["host"] + ref.path
	}
	if !strings.HasPrefix(ref.path, "/") {
		ref.path = "/" + ref.path
	}
	if strings.HasSuffix(ref.path, "/") {
		ref.path = ref.path[:len(ref.path)-1]
	}
	ref.name = ref.path[strings.LastIndex(ref.path, "/")+1:]
	ref.user = loc["user"]
	if ref.user == "" {
		ref.user = os.Getenv("COPY_USERNAME")
	}
	ref.password = loc["password"]
	if ref.password == "" {
		ref.password = os.Getenv("COPY_PASSWORD")
	}
	return ref
}

func (ccsp *CopyCloudStorageProvider) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("X-Client-Type", "api")
	req.Header.Set("X-Api-Version", "1.0")
	if req.Header.Get("X-Authorization") == "" {
		req.Header.Set("X-Authorization", "")
	}
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return http.DefaultClient.Do(req)
}

func (ccsp *CopyCloudStorageProvider) post(method string, params interface{}, headers http.Header) (map[string]interface{}, error) {
	bs, _ := json.Marshal(params)
	vs := url.Values{}
	vs.Add("data", string(bs))
	req, err := http.NewRequest("POST", "https://api.copy.com/"+method, strings.NewReader(vs.Encode()))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header[k] = v
	}
	res, err := ccsp.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bs, _ = ioutil.ReadAll(res.Body)
	var obj map[string]interface{}
	err = json.Unmarshal(bs, &obj)
	if err != nil {
		return nil, fmt.Errorf("invalid json: %v", string(bs))
	}
	e, ok := obj["error_string"]
	if ok {
		return nil, fmt.Errorf("%s", e)
	}
	return obj, nil
}

func (ccsp *CopyCloudStorageProvider) updateObjects(params interface{}, tok string) (map[string]interface{}, error) {
	h := make(http.Header)
	h.Set("X-Authorization", tok)
	log.Println(params)
	return ccsp.post("update_objects", map[string]interface{}{
		"meta": []interface{}{params},
	}, h)
}

func (ccsp *CopyCloudStorageProvider) auth(username, password string) (string, error) {
	obj, err := ccsp.post("auth_user", map[string]string{"username": username, "password": password}, nil)
	if err != nil {
		return "", err
	}
	at, ok := obj["auth_token"].(string)
	if !ok || at == "" {
		return "", fmt.Errorf("invalid login")
	}
	return at, nil
}

func (ccsp *CopyCloudStorageProvider) Delete(loc Location) error {
	ref := ccsp.ref(loc)
	tok, err := ccsp.auth(ref.user, ref.password)
	if err != nil {
		return err
	}
	_, err = ccsp.updateObjects(map[string]string{"action": "remove", "path": ref.path}, tok)
	return err
}

func (ccsp *CopyCloudStorageProvider) Get(loc Location) (io.ReadCloser, error) {
	ref := ccsp.ref(loc)
	tok, err := ccsp.auth(ref.user, ref.password)
	if err != nil {
		return nil, err
	}
	bs, _ := json.Marshal(map[string]string{"path": ref.path})
	vs := url.Values{}
	vs.Add("data", string(bs))
	req, err := http.NewRequest("POST", "https://api.copy.com/download_object", strings.NewReader(vs.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Authorization", tok)
	res, err := ccsp.do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode/100 == 2 {
		return res.Body, nil
	}
	res.Body.Close()
	return nil, fmt.Errorf("%s", res.Status)
}

func (ccsp *CopyCloudStorageProvider) Put(loc Location, rdr io.Reader) error {
	ref := ccsp.ref(loc)
	tok, err := ccsp.auth(ref.user, ref.password)
	if err != nil {
		return err
	}

	headers := map[string]string{
		"X-Client-Type":    "API",
		"X-Client-Version": "1.0.00",
		"Accept":           "application/json",
		"X-Api-Version":    "1.0",
		"X-Authorization":  tok,
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	w := multipart.NewWriter(pw)
	go func() {
		defer w.Close()
		defer pw.Close()
		for k, v := range headers {
			fw, err := w.CreateFormField(k)
			if err != nil {
				return
			}
			fw.Write([]byte(v))
		}
		fw, err := w.CreateFormFile("files[]", ref.name)
		if err != nil {
			return
		}
		io.Copy(fw, rdr)
	}()

	req, err := http.NewRequest("POST", "https://apiweb.copy.com/rest/meta/copy/stack", pr)
	if err != nil {
		return err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := ccsp.do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode/100 == 2 {
		return nil
	}
	return fmt.Errorf("%s", res.Status)
}

func (ccsp *CopyCloudStorageProvider) List(loc Location) ([]string, error) {
	ref := ccsp.ref(loc)
	tok, err := ccsp.auth(ref.user, ref.password)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", "https://api.copy.com/rest/meta/copy"+ref.path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Authorization", tok)
	res, err := ccsp.do(req)
	if err != nil {
		return nil, err
	}
	var obj map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&obj)
	if err != nil {
		return nil, err
	}
	children, ok := obj["children"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected children, got %T", obj["children"])
	}
	arr := make([]string, 0, len(children))
	for _, child := range children {
		m, ok := child.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected child to be object, got %T", child)
		}
		p, ok := m["path"].(string)
		if !ok {
			return nil, fmt.Errorf("expected path")
		}
		if len(p) > len(ref.path) {
			p = p[len(ref.path)+1:]
		}
		arr = append(arr, p)
	}

	return arr, nil
}

func (ccsp *CopyCloudStorageProvider) Version(loc Location, previous string) (string, error) {
	panic("not implemented")
}

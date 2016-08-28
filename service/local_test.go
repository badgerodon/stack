package service

import (
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalServiceManager(t *testing.T) {
	assert := assert.New(t)

	LocalServiceManagerTickerDuration = time.Millisecond * 500

	writeFilePath := filepath.Join(os.TempDir(), uuid.New()+".txt")
	defer os.Remove(writeFilePath)

	batchFilePath := filepath.Join(os.TempDir(), uuid.New()+".bat")
	defer os.Remove(batchFilePath)

	ioutil.WriteFile(batchFilePath, []byte(`
echo %time% > `+writeFilePath+`
`), 0644)

	stateFilePath := filepath.Join(os.TempDir(), uuid.New())
	defer os.Remove(stateFilePath)

	lsm := NewLocalServiceManager(stateFilePath)
	err := lsm.Start()
	assert.Nil(err)

	svc := Service{
		Name:      "fake-service",
		Directory: os.TempDir(),
		Command:   []string{"cmd.exe", "/q", "/c", batchFilePath},
	}
	err = lsm.Install(svc)
	assert.Nil(err)

	var foundFirst, foundLast string
	fiveSecondsFromNow := time.Now().Add(5 * time.Second)
	for time.Now().Before(fiveSecondsFromNow) {
		bs, err := ioutil.ReadFile(writeFilePath)
		if err == nil {
			if foundFirst == "" {
				foundFirst = string(bs)
			}
			foundLast = string(bs)

			if foundLast != foundFirst {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	assert.NotEmpty(foundFirst, "expected service manager to run command and command to write file")
	assert.NotEqual(foundLast, foundFirst, "expected service manager to restart commands when they stop")

	err = lsm.Uninstall(svc.Name)
	assert.Nil(err)
}

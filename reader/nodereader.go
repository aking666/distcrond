package reader

import (
	"os"
	"errors"
	"fmt"
	"strings"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"github.com/martin-helmich/distcrond/domain"
	"github.com/martin-helmich/distcrond/logging"
)

type NodeReader struct {
	receiver NodeReceiver
}

type NodeReceiver interface {
	AddNode(domain.Node)
}

func NewNodeReader(receiver NodeReceiver) *NodeReader {
	reader := new(NodeReader)
	reader.receiver = receiver

	return reader
}

func (r NodeReader) ReadFromDirectory(directory string) error {
	logging.Info("Reading node configuration")

	var walk filepath.WalkFunc = func(path string, file os.FileInfo, err error) error {
		if file.IsDir() {
			return nil
		}

		if file.Name()[0] == '.' {
			logging.Debug("Skipping %s", path)
			return nil
		}

		fileContents, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		wrapError := func(err error) error {
			return errors.New(fmt.Sprintf("Error parsing file %s: %s", file.Name(), err))
		}

		node := domain.Node{
			Name: strings.Replace(file.Name(), ".json", "", 1),
		}
		err = json.Unmarshal(fileContents, &node)
		if err != nil {
			return wrapError(err)
		}

		logging.Debug("Read node from %s: %s", file.Name(), node)

		if validErr := node.IsValid(); validErr != nil {
			return wrapError(validErr)
		}

		r.receiver.AddNode(node)

		return nil
	}

	if err := filepath.Walk(directory, walk); err != nil {
		return err
	}

	return nil
}
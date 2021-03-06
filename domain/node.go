package domain

import (
	"errors"
	"fmt"
	"sync"
)

const (
	CONN_LOCAL = "local"
	CONN_SSH = "ssh"
)

const (
	STATUS_UP = iota
	STATUS_DOWN
)

type ConnectionType string
type NodeStatus int

type ConnectionOptions struct {
	SshHost string `json:"ssh_host"`
	SshUser string `json:"ssh_user"`
	SshKeyFile string `json:"ssh_private_key_file"`
}

func (o ConnectionOptions) SetDefaults(forType string) {
	if forType == CONN_SSH {
		o.SshHost = "localhost"
		o.SshUser = "root"
		o.SshKeyFile = "~/.ssh/id_rsa"
	}
}

func (o ConnectionOptions) IsValid(forType ConnectionType) error {
	switch {
	case forType == CONN_SSH:
		if len(o.SshHost) == 0 {
			return errors.New("SSH host is empty")
		}

		if len(o.SshUser) == 0 {
			return errors.New("SSH user is empty")
		}

		if len(o.SshKeyFile) == 0 {
			return errors.New("SSH key is empty")
		}

	case forType == CONN_LOCAL:
		return nil
	}

	return nil
}

type ExecutionStrategy interface {
	HealthCheck() error
	ExecuteCommand(job *Job, report *RunReportItem) error
}

type NodeJson struct {
	Name string `json:"name"`
	Roles []string `json:"roles"`
	ConnectionType string `json:"connection_type"`
	ConnectionOptions ConnectionOptions `json:"connection_options"`
}

type Node struct {
	Name              string
	Roles             []string
	ConnectionType    ConnectionType
	ConnectionOptions ConnectionOptions
	Status            NodeStatus
	RunningJobs       int32

	ExecutionStrategy ExecutionStrategy
	Lock              sync.RWMutex
}

func NewNodeFromJson(name string, json NodeJson) (Node, error) {
	node := Node{}
	node.Name = name
	node.Roles = json.Roles
	node.ConnectionType = ConnectionType(json.ConnectionType)
	node.ConnectionOptions = json.ConnectionOptions

	return node, nil
}

func (n Node) IsValid() error {
	if n.ConnectionType != CONN_LOCAL && n.ConnectionType != CONN_SSH {
		return errors.New("Invalid connection type (must be either " + CONN_LOCAL + " or " + CONN_SSH + ").")
	}

	if len(n.Name) == 0 {
		return errors.New("Name is empty")
	}

	if err := n.ConnectionOptions.IsValid(n.ConnectionType); err != nil {
		return errors.New(fmt.Sprintf("Invalid connection options: %s", err))
	}

	return nil
}

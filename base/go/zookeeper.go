package base

import (
	"fmt"
	"log"
	"time"

	"github.com/go-zookeeper/zk"
)

// ZookeeperConfig stores the configuration for the Zookeeper connection.
type ZookeeperConfig struct {
	Address []string
	Timeout time.Duration
}

// NewZookeeperClient creates a new Zookeeper client.
func NewZookeeperClient(config ZookeeperConfig) (*zk.Conn, error) {
	conn, _, err := zk.Connect(config.Address, config.Timeout)
	if err != nil {
		log.Fatalf("failed to connect to Zookeeper: %v", err)
		return nil, err
	}

	fmt.Println("Connected to Zookeeper!")
	return conn, nil
}

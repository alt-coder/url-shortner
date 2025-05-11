package service

import (
	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/mock"
)

// MockZookeeperClient is a mock for ZkClientInterface
type MockZookeeperClient struct {
	mock.Mock
}

var _ ZkClientInterface = (*MockZookeeperClient)(nil)

func (m *MockZookeeperClient) Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error) {
	args := m.Called(path, data, flags, acl)
	return args.String(0), args.Error(1)
}

func (m *MockZookeeperClient) Get(path string) ([]byte, *zk.Stat, error) {
	args := m.Called(path)
	var stat *zk.Stat
	if args.Get(1) != nil {
		stat = args.Get(1).(*zk.Stat)
	}
	return args.Get(0).([]byte), stat, args.Error(2)
}

func (m *MockZookeeperClient) Set(path string, data []byte, version int32) (*zk.Stat, error) {
	args := m.Called(path, data, version)
	var stat *zk.Stat
	if args.Get(0) != nil {
		stat = args.Get(0).(*zk.Stat)
	}
	return stat, args.Error(1)
}

func (m *MockZookeeperClient) Exists(path string) (bool, *zk.Stat, error) {
	args := m.Called(path)
	var stat *zk.Stat
	if args.Get(1) != nil {
		stat = args.Get(1).(*zk.Stat)
	}
	return args.Bool(0), stat, args.Error(2)
}

func (m *MockZookeeperClient) Close() {
	m.Called()
}
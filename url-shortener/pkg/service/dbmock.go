package service



import (
	"github.com/alt-coder/url-shortener/url-shortener/pkg/dataModel"
	
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock for dataModel.DataAccessLayer
type MockDB struct {
	mock.Mock
}

// Ensure MockDB implements dataModel.DataAccessLayer
var _ dataModel.DataAccessLayer = (*MockDB)(nil)


func (m *MockDB) AutoMigrate(dst ...interface{}) error {
	args := m.Called(dst)
	return args.Error(0)
}

func (m *MockDB) CreateURLMapping(urlMapping *dataModel.URLMapping) error {
	args := m.Called(urlMapping)
	return args.Error(0)
}

func (m *MockDB) GetLongURL(shortURLID string) (string, error) {
	args := m.Called(shortURLID)
	return args.String(0), args.Error(1)
}

func (m *MockDB) CreateUser(user *dataModel.User) error {
	args := m.Called(user)
	if args.Error(0) == nil {
		user.ID = 1 // Simulate GORM behavior
		user.APIKey = uuid.New() // APIKey is uuid.UUID
	}
	return args.Error(0)
}

func (m *MockDB) GetUserByEmail(email string) (*dataModel.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dataModel.User), args.Error(1)
}

func (m *MockDB) GetAPIKeyByEmail(email string) (string, error) {
	args := m.Called(email)
	return args.String(0), args.Error(1)
}

func (m *MockDB) CheckAPIKey(apiKey string) (bool, error) {
	args := m.Called(apiKey)
	return args.Bool(0), args.Error(1)
}

func (m* MockDB) GetTopDomains(limit int) ([]dataModel.DomainCount, error){
	args := m.Called(limit)
	return args.Get(0).([]dataModel.DomainCount), args.Error(1)
}
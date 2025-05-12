package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/alt-coder/url-shortener/url-shortener/pkg/dataModel"
	proto "github.com/alt-coder/url-shortener/url-shortener/proto"
	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	ZkCounterPath       = "/counter"
	ZkInitialCounterVal = "0"
)

func TestNewServer(t *testing.T) {
	os.Setenv(GrpcPort, "50051")
	os.Setenv(HttpPort, "8080")
	os.Setenv(PostgresHost, "localhost")
	os.Setenv(PostgresPort, "5432")
	os.Setenv(PostgresUser, "user")
	os.Setenv(PostgresPassword, "password")
	os.Setenv(PostgresDBName, "testdb")
	os.Setenv(RedisHost, "localhost")
	os.Setenv(RedisPort, "6379")
	os.Setenv(RedisPassword, "")
	os.Setenv(ZookeeperHost, "localhost")
	os.Setenv(ZookeeperPort, "2181")

	t.Run("Invalid Postgres Port", func(t *testing.T) {
		originalPort := os.Getenv(PostgresPort)
		os.Setenv(PostgresPort, "invalid")
		defer os.Setenv(PostgresPort, originalPort)
		_, err := NewUrlShortnerService() // NewServer uses base.New... clients
		assert.Error(t, err, "Expected error for invalid port")
	})
}

func TestShortenURL(t *testing.T) {
	mockDb := new(MockDB)
	mockZk := new(MockZookeeperClient)

	s := &UrlShortenerService{
		Config: Config{GrpcPort: "50051", HttpPort: "8080"},
		db:     mockDb,

		ZookeeperClient: mockZk, // This line makes the test work IF service uses interface
		RedisClient:     nil,
		mu:              sync.Mutex{},
	}

	ctx := context.Background()

	t.Run("Successful ShortenURL", func(t *testing.T) {
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = false // Reset state
		req := &proto.ShortenURLRequest{
			ApiKey:  "valid-api-key",
			LongUrl: "http://example.com/very/long/url",
		}

		mockDb.On("CheckAPIKey", "valid-api-key").Return(true, nil).Once()
		mockDb.On("CreateURLMapping", mock.AnythingOfType("*dataModel.URLMapping")).Run(func(args mock.Arguments) {
		}).Return(nil).Once()
		requestCounterFunc = func(s *UrlShortenerService) (int64, error) {
			return 12345, nil
		}

		resp, err := s.ShortenURL(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotEmpty(t, resp.ShortUrl)
		assert.Equal(t, base62Encode(12345), resp.ShortUrl)

		mockDb.AssertExpectations(t)
		mockZk.AssertExpectations(t)
	})

	t.Run("Missing API Key", func(t *testing.T) {
		// s (service) is reused, state is implicitly carried unless reset
		req := &proto.ShortenURLRequest{ApiKey: "", LongUrl: "http://example.com/another/url"}
		resp, err := s.ShortenURL(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, ErrMissingApiKey, err)
		assert.Nil(t, resp)
	})

	t.Run("Invalid API Key", func(t *testing.T) {
		req := &proto.ShortenURLRequest{ApiKey: "invalid-api-key", LongUrl: "http://example.com/another/url"}
		mockDb.On("CheckAPIKey", "invalid-api-key").Return(false, nil).Once()
		resp, err := s.ShortenURL(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidApiKey, err)
		assert.Nil(t, resp)
		mockDb.AssertExpectations(t)
	})

	t.Run("DB Error on CheckAPIKey", func(t *testing.T) {
		req := &proto.ShortenURLRequest{ApiKey: "any-api-key", LongUrl: "http://example.com/some/url"}
		dbErr := errors.New("db error checking api key")
		mockDb.On("CheckAPIKey", "any-api-key").Return(false, dbErr).Once()
		resp, err := s.ShortenURL(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, resp)
		mockDb.AssertExpectations(t)
	})

	t.Run("DB Error on CreateURLMapping", func(t *testing.T) {
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = false
		req := &proto.ShortenURLRequest{ApiKey: "valid-api-key", LongUrl: "http://example.com/long/url"}
		mockDb.On("CheckAPIKey", "valid-api-key").Return(true, nil).Once()
		requestCounterFunc = func(s *UrlShortenerService) (int64, error) {
			return 12345, nil
		}

		dbErr := errors.New("db error creating url mapping")
		mockDb.On("CreateURLMapping", mock.AnythingOfType("*dataModel.URLMapping")).Return(dbErr).Once()

		resp, err := s.ShortenURL(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, resp)
		mockDb.AssertExpectations(t)
		mockZk.AssertExpectations(t)
	})

	t.Run("RequestCounter Error", func(t *testing.T) {
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = false // Reset state
		req := &proto.ShortenURLRequest{
			ApiKey:  "valid-api-key",
			LongUrl: "http://example.com/another/long/url",
		}
		mockDb.On("CheckAPIKey", "valid-api-key").Return(true, nil).Once()

		// Simulate an error from Zookeeper Exits
		requestCounterFunc = func(s *UrlShortenerService) (int64, error) {
			return -1, fmt.Errorf("zookeeper Exist error")
		}
		resp, err := s.ShortenURL(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zookeeper Exist error")
		assert.Nil(t, resp)

		mockDb.AssertExpectations(t)
		mockZk.AssertExpectations(t)
	})
}

func TestGetURL(t *testing.T) {
	mockDb := new(MockDB)
	s := &UrlShortenerService{db: mockDb} // ZK and Redis not directly used by GetURL
	ctx := context.Background()

	t.Run("Successful GetURL", func(t *testing.T) {
		req := &proto.GetURLRequest{ShortUrl: "testShort"}
		expectedLongURL := "http://example.com/original/long/url"
		mockDb.On("GetLongURL", "testShort").Return(expectedLongURL, nil).Once()
		resp, err := s.GetURL(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedLongURL, resp.LongUrl)
		mockDb.AssertExpectations(t)
	})

	t.Run("GetURL DB Error", func(t *testing.T) {
		req := &proto.GetURLRequest{ShortUrl: "testShort"}
		dbErr := errors.New("db error getting long url")
		mockDb.On("GetLongURL", "testShort").Return("", dbErr).Once()
		resp, err := s.GetURL(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, resp)
		mockDb.AssertExpectations(t)
	})
}

func TestCreateUser(t *testing.T) {
	mockDb := new(MockDB)
	s := &UrlShortenerService{db: mockDb}
	ctx := context.Background()

	t.Run("Successful CreateUser", func(t *testing.T) {
		req := &proto.CreateUserRequest{FirstName: "Test", LastName: "User", Email: "test.user@example.com"}

		var capturedUser *dataModel.User
		mockDb.On("CreateUser", mock.MatchedBy(func(user *dataModel.User) bool {
			capturedUser = user
			return user.Email == req.Email
		})).Return(nil).Once()

		resp, err := s.CreateUser(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		// User ID and APIKey are set by the mock. Default mock sets ID to 1 and generates a new UUID.
		assert.Equal(t, "1", resp.UserId)
		// Check if API key in response matches the one set on the captured user by the mock
		assert.Equal(t, capturedUser.APIKey.String(), resp.ApiKey)
		mockDb.AssertExpectations(t)
	})

	t.Run("CreateUser DB Error", func(t *testing.T) {
		req := &proto.CreateUserRequest{FirstName: "Error", LastName: "Case", Email: "error.case@example.com"}
		dbErr := errors.New("db error creating user")

		mockDb.On("CreateUser", mock.MatchedBy(func(user *dataModel.User) bool {
			return user.Email == req.Email
		})).Return(dbErr).Once()

		resp, err := s.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, resp)
		mockDb.AssertExpectations(t)
	})
}

func TestFetchApiKey(t *testing.T) {
	mockDb := new(MockDB)
	s := &UrlShortenerService{db: mockDb}
	ctx := context.Background()

	t.Run("Successful FetchApiKey", func(t *testing.T) {
		req := &proto.FetchApiKeyRequest{Email: "test.user@example.com"}
		expectedAPIKey := uuid.New().String()
		mockDb.On("GetAPIKeyByEmail", "test.user@example.com").Return(expectedAPIKey, nil).Once()
		resp, err := s.FetchApiKey(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedAPIKey, resp.ApiKey)
		mockDb.AssertExpectations(t)
	})

	t.Run("FetchApiKey DB Error", func(t *testing.T) {
		req := &proto.FetchApiKeyRequest{Email: "error.user@example.com"}
		dbErr := errors.New("db error fetching api key")
		mockDb.On("GetAPIKeyByEmail", "error.user@example.com").Return("", dbErr).Once()
		resp, err := s.FetchApiKey(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, resp)
		mockDb.AssertExpectations(t)
	})
}

func TestRedirectHandler(t *testing.T) {
	mockDb := new(MockDB)
	s := &UrlShortenerService{db: mockDb}

	t.Run("Successful Redirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/d/testShort", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"shortChar": "testShort"})
		expectedLongURL := "http://example.com/redirected/url"
		mockDb.On("GetLongURL", "testShort").Return(expectedLongURL, nil).Once()
		http.HandlerFunc(s.redirectHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusFound, rr.Code)
		assert.Equal(t, expectedLongURL, rr.Header().Get("Location"))
		mockDb.AssertExpectations(t)
	})

	t.Run("RedirectHandler GetURL Error", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/d/errorShort", nil)
		rr := httptest.NewRecorder()
		req = mux.SetURLVars(req, map[string]string{"shortChar": "errorShort"})
		dbErr := errors.New("db error for GetURL in redirect")
		mockDb.On("GetLongURL", "errorShort").Return("", dbErr).Once()
		http.HandlerFunc(s.redirectHandler).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockDb.AssertExpectations(t)
	})
}

func TestRequestCounter(t *testing.T) {

	t.Run("Counter increments within current range", func(t *testing.T) {
		mockZk := new(MockZookeeperClient) // Isolated mock
		s := &UrlShortenerService{         // Isolated service instance
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 5
		s.uppLimitVal = 10
		s.isCounterExists = true
		counter, err := s.requestCounter()
		assert.NoError(t, err)
		assert.Equal(t, int64(6), counter)
		mockZk.AssertExpectations(t) // Assert expectations for this sub-test's mock
	})

	t.Run("Counter hits upper limit, fetches from ZK", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 10 // Corrected: to make s.currentCounterVal >= s.uppLimitVal
		s.uppLimitVal = 10
		s.isCounterExists = true

		initialZkCounter := int64(1000)
		valueToSetInZk := initialZkCounter + 10000 // CounterRange is 10000

		mockZk.On("Get", ZkCounterPath).Return([]byte(strconv.FormatInt(initialZkCounter, 10)), &zk.Stat{Version: 5}, nil).Once()
		mockZk.On("Set", ZkCounterPath, []byte(strconv.FormatInt(valueToSetInZk, 10)), int32(5)).Return(&zk.Stat{}, nil).Once()

		counter, err := s.requestCounter()
		assert.NoError(t, err)
		assert.Equal(t, initialZkCounter+1, counter) // Service reads 1000, increments to 1001
		mockZk.AssertExpectations(t)
	})

	t.Run("Counter hits upper limit, ZK node does not exist, creates node", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = false

		mockZk.On("Exists", ZkCounterPath).Return(false, (*zk.Stat)(nil), nil).Once()
		mockZk.On("Create", ZkCounterPath, []byte(ZkInitialCounterVal), int32(0), zk.WorldACL(zk.PermAll)).Return(ZkCounterPath, nil).Once()

		initialZkCounterAfterCreate, _ := strconv.ParseInt(ZkInitialCounterVal, 10, 64) // This is 0
		valueToSetInZk := initialZkCounterAfterCreate + 10000                           // 0 + 10000 = 10000

		mockZk.On("Get", ZkCounterPath).Return([]byte(ZkInitialCounterVal), &zk.Stat{Version: 0}, nil).Once()                   // Get returns "0"
		mockZk.On("Set", ZkCounterPath, []byte(strconv.FormatInt(valueToSetInZk, 10)), int32(0)).Return(&zk.Stat{}, nil).Once() // Set "10000"

		counter, err := s.requestCounter()
		assert.NoError(t, err)
		assert.Equal(t, initialZkCounterAfterCreate+1, counter) // Service reads 0, increments to 1
		assert.True(t, s.isCounterExists)
		mockZk.AssertExpectations(t)
	})

	t.Run("Error during ZK Get", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = true // Force ZK Get path, assuming checkZkCounter would have set this.
		// If checkZkCounter was meant to run and fail before Get, this needs adjustment.
		// For now, focusing on Get failure itself.

		zkErr := errors.New("zk get error")
		mockZk.On("Get", ZkCounterPath).Return([]byte{}, &zk.Stat{}, zkErr).Once()
		_, err := s.requestCounter()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "zk get error")
		}
		mockZk.AssertExpectations(t)
	})

	t.Run("Error during strconv.Atoi", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = true

		mockZk.On("Get", ZkCounterPath).Return([]byte("not-a-number"), &zk.Stat{Version: 1}, nil).Once()
		_, err := s.requestCounter()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "not-a-number")
		}
		mockZk.AssertExpectations(t)
	})

	t.Run("Error during ZK Set", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = true

		initialZkCounter := int64(50)
		zkSetErr := errors.New("zk set error")
		mockZk.On("Get", ZkCounterPath).Return([]byte(strconv.FormatInt(initialZkCounter, 10)), &zk.Stat{Version: 1}, nil).Once()
		mockZk.On("Set", ZkCounterPath, []byte(strconv.FormatInt(initialZkCounter+10000, 10)), int32(1)).Return(nil, zkSetErr).Once()
		_, err := s.requestCounter()
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "zk set error")
		}
		mockZk.AssertExpectations(t)
	})

	t.Run("Error from checkZkCounter - Exists fails", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = false

		zkErr := errors.New("zk exists error")
		mockZk.On("Exists", ZkCounterPath).Return(false, nil, zkErr).Once()
		_, err := s.requestCounter() // This will call checkZkCounter
		assert.Error(t, err)
		if err != nil {
			assert.Contains(t, err.Error(), "zk exists error")
		}
		mockZk.AssertExpectations(t)
	})

	t.Run("Error from checkZkCounter - Create fails", func(t *testing.T) {
		mockZk := new(MockZookeeperClient)
		s := &UrlShortenerService{
			ZookeeperClient: mockZk,
			mu:              sync.Mutex{},
		}
		s.currentCounterVal = 0
		s.uppLimitVal = 0
		s.isCounterExists = false

		zkErr := errors.New("zk create error")
		mockZk.On("Exists", ZkCounterPath).Return(false, (*zk.Stat)(nil), nil).Once()
		mockZk.On("Create", ZkCounterPath, []byte(ZkInitialCounterVal), int32(0), zk.WorldACL(zk.PermAll)).Return("", zkErr).Once()
		_, err := s.requestCounter()
		assert.Error(t, err)
		if err != nil {
			// Similar to Exists fails, checkZkCounter returns fmt.Errorf("failed to create counter node: %w", err)
			assert.Contains(t, err.Error(), "create")
		}
		mockZk.AssertExpectations(t)
	})
}

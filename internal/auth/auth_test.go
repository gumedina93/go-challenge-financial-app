package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go-challenge-financial-chat/internal/models"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) CreateUser(username, passwordHash string) error {
	args := m.Called(username, passwordHash)
	return args.Error(0)
}

func (m *MockDB) GetUser(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockDB) SaveMessage(userID int, username, content string) error {
	args := m.Called(userID, username, content)
	return args.Error(0)
}

func (m *MockDB) GetRecentMessages(limit int) ([]models.Message, error) {
	args := m.Called(limit)
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestService_Register(t *testing.T) {
	mockDB := new(MockDB)
	service := NewService(mockDB)

	tests := []struct {
		name          string
		username      string
		password      string
		mockSetup     func()
		expectedError bool
	}{
		{
			name:     "Valid registration",
			username: "testuser",
			password: "password123",
			mockSetup: func() {
				mockDB.On("CreateUser", "testuser", mock.Anything).Return(nil)
			},
			expectedError: false,
		},
		{
			name:          "Username too short",
			username:      "te",
			password:      "password123",
			mockSetup:     func() {},
			expectedError: true,
		},
		{
			name:          "Password too short",
			username:      "testuser",
			password:      "pass",
			mockSetup:     func() {},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			err := service.Register(tt.username, tt.password)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	mockDB := new(MockDB)
	service := NewService(mockDB)

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
	}

	tests := []struct {
		name          string
		username      string
		password      string
		mockSetup     func()
		expectedError bool
	}{
		{
			name:     "Valid login",
			username: "testuser",
			password: password,
			mockSetup: func() {
				mockDB.On("GetUser", "testuser").Return(testUser, nil)
			},
			expectedError: false,
		},
		{
			name:     "Invalid password",
			username: "testuser",
			password: "wrongpassword",
			mockSetup: func() {
				mockDB.On("GetUser", "testuser").Return(testUser, nil)
			},
			expectedError: true,
		},
		{
			name:     "User not found",
			username: "nonexistent",
			password: password,
			mockSetup: func() {
				mockDB.On("GetUser", "nonexistent").Return(nil, errors.New("user not fund"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			user, err := service.Login(tt.username, tt.password)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.username, user.Username)
			}
		})
	}
}

func TestService_SessionManagement(t *testing.T) {
	mockDB := new(MockDB)
	service := NewService(mockDB)

	t.Run("SetSession", func(t *testing.T) {
		w := httptest.NewRecorder()
		service.SetSession(w, "testuser")

		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "session", cookies[0].Name)
		assert.Equal(t, "testuser", cookies[0].Value)
		assert.Equal(t, 3600*24, cookies[0].MaxAge)
	})

	t.Run("GetSession", func(t *testing.T) {
		// Create a request with a session cookie
		req := httptest.NewRequest("GET", "/", nil)
		cookie := &http.Cookie{
			Name:  "session",
			Value: "testuser",
		}
		req.AddCookie(cookie)

		username, err := service.GetSession(req)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", username)
	})

	t.Run("GetSession_NoCookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		username, err := service.GetSession(req)
		assert.Error(t, err)
		assert.Empty(t, username)
	})

	t.Run("ClearSession", func(t *testing.T) {
		w := httptest.NewRecorder()
		service.ClearSession(w)

		cookies := w.Result().Cookies()
		assert.Len(t, cookies, 1)
		assert.Equal(t, "session", cookies[0].Name)
		assert.Empty(t, cookies[0].Value)
		assert.Equal(t, -1, cookies[0].MaxAge)
	})
}

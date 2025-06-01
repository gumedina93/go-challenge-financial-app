package auth

import (
	"errors"
	"go-challenge-financial-chat/internal/database"
	"go-challenge-financial-chat/internal/models"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type Service struct {
	db database.Database
}

func NewService(db database.Database) *Service {
	return &Service{db: db}
}

func (s *Service) Register(username, password string) error {
	if len(username) < 3 || len(password) < 6 {
		return errors.New("username must be at least 3 characters and password at least 6 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.db.CreateUser(username, string(hashedPassword))
}

func (s *Service) Login(username, password string) (*models.User, error) {
	user, err := s.db.GetUser(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

func (s *Service) SetSession(w http.ResponseWriter, username string) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    username,
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
}

func (s *Service) GetSession(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func (s *Service) ClearSession(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

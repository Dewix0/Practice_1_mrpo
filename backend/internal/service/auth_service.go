package service

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"shoe-store/internal/model"
	"shoe-store/internal/repository"
)

type AuthService struct {
	UserRepo  *repository.UserRepo
	JwtSecret string
}

func NewAuthService(repo *repository.UserRepo) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-key" //nolint:gosec
	}
	return &AuthService{
		UserRepo:  repo,
		JwtSecret: secret,
	}
}

// Login authenticates the user and returns a LoginResponse containing a JWT token and user info.
func (s *AuthService) Login(login, password string) (*model.LoginResponse, error) {
	user, err := s.UserRepo.FindByLogin(login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid login or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid login or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.ID,
		"role":   user.RoleName,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.JwtSecret))
	if err != nil {
		return nil, err
	}

	// Clear password before returning
	user.Password = ""

	return &model.LoginResponse{
		Token: tokenString,
		User:  *user,
	}, nil
}

// GetUserByID returns a user by ID (for the /me endpoint).
func (s *AuthService) GetUserByID(id int64) (*model.User, error) {
	user, err := s.UserRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	user.Password = ""
	return user, nil
}

package service

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"task-manager-api/apperrors"
	"task-manager-api/models"
)

type UserService struct {
	UserRepo UserRepository
}
type UserRepository interface {
	RegisterUser(email string, hashedPassword string) error
	GetUserByEmail(email string) (models.User, error)
}

func (S *UserService) RegisterUser(email string, password string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return apperrors.ErrEmptyEmail
	}
	if password == "" {
		return apperrors.ErrEmptyPassword
	}
	if len(password) < 6 {
		return apperrors.PasswordTooShort
	}
	hasedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return S.UserRepo.RegisterUser(email, string(hasedPassword))
}
func (S *UserService) LoginUser(email string, password string) (models.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return models.User{}, apperrors.ErrEmptyEmail
	}
	if password == "" {
		return models.User{}, apperrors.ErrEmptyPassword
	}
	if len(password) < 6 {
		return models.User{}, apperrors.PasswordTooShort
	}
	user, err := S.UserRepo.GetUserByEmail(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, apperrors.ErrInvalidCredentials
		}
		return models.User{}, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return models.User{}, apperrors.ErrInvalidCredentials
	}
	return user, nil
}

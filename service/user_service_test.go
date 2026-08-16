package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"task-manager-api/apperrors"
	"task-manager-api/models"
	"testing"
)

type FakeUserRepository struct {
	RegisterUserCalled   bool
	RegisterUserEmail    string
	RegisterUserPassword string
	RegisterUserError    error

	LoginUserCalled      bool
	LoginUserEmail       string
	GetUserByEmailCalled bool
	GetUserByEmailError  error
	GetUserByEmailUser   models.User
}

func (f *FakeUserRepository) RegisterUser(email string, hashedPassword string) error {
	f.RegisterUserCalled = true
	f.RegisterUserEmail = email
	f.RegisterUserPassword = hashedPassword
	if f.RegisterUserError != nil {
		return f.RegisterUserError
	}
	return nil
}
func (f *FakeUserRepository) GetUserByEmail(email string) (models.User, error) {
	f.GetUserByEmailCalled = true
	if f.GetUserByEmailError != nil {
		return models.User{}, f.GetUserByEmailError
	}
	return f.GetUserByEmailUser, nil
}
func TestRegisterUser_EmptyEmail(t *testing.T) {
	fakeRepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakeRepo}
	err := service.RegisterUser("   ", "password")
	if err != apperrors.ErrEmptyEmail {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}

func TestRegisterUser_EmptyPassword(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakerepo}
	err := service.RegisterUser("email@gmail.com", "")
	if !errors.Is(err, apperrors.ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}
func TestRegisterUser_PasswordTooShort(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakerepo}
	err := service.RegisterUser("email@gmail.com", "pass")
	if !errors.Is(err, apperrors.PasswordTooShort) {
		t.Errorf("expected PasswordTooShort, got %v", err)
	}
}
func TestRegisterUser_Success(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakerepo}
	err := service.RegisterUser("email@gmail.com", "password")
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if fakerepo.RegisterUserEmail != "email@gmail.com" {
		t.Errorf("expected email same but got %s", fakerepo.RegisterUserEmail)
	}
	err = bcrypt.CompareHashAndPassword([]byte(fakerepo.RegisterUserPassword), []byte("password"))
	if err != nil {
		t.Errorf("password was not hashed correctly")
	}
	if !fakerepo.RegisterUserCalled {
		t.Errorf("expected RegisterUserCalled to be true")
	}

}

func TestRegisterUser_EmailExists(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	fakerepo.RegisterUserError = apperrors.ExistEmailError
	service := UserService{UserRepo: fakerepo}
	err := service.RegisterUser("email@gmail.com", "password")
	if !errors.Is(err, apperrors.ExistEmailError) {
		t.Errorf("expected ExistEmailError, got %v", err)
	}
}
func TestRegisterUser_RepositoryError(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	fakerepo.RegisterUserError = errors.New("database error")
	service := UserService{UserRepo: fakerepo}
	err := service.RegisterUser("email@gmail.com", "password")
	if !errors.Is(err, fakerepo.RegisterUserError) {
		t.Errorf("expected database error got %v", err)
	}
}

func TestLoginUser_EmptyEmail(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakerepo}
	_, err := service.LoginUser("   ", "password")
	if !errors.Is(err, apperrors.ErrEmptyEmail) {
		t.Errorf("expected ErrEmptyEmail, got %v", err)
	}
}
func TestLoginUser_EmptyPassword(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakerepo}
	_, err := service.LoginUser("email@gmail.com", "")
	if !errors.Is(err, apperrors.ErrEmptyPassword) {
		t.Errorf("expected ErrEmptyPassword, got %v", err)
	}
}
func TestLoginUser_PasswordTooShort(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	service := UserService{UserRepo: fakerepo}
	_, err := service.LoginUser("email@gmail.com", "pass")
	if !errors.Is(err, apperrors.PasswordTooShort) {
		t.Errorf("expected PasswordTooShort, got %v", err)
	}
}
func TestLoginUser_UserNotFound(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	fakerepo.GetUserByEmailError = apperrors.ErrInvalidCredentials
	service := UserService{UserRepo: fakerepo}
	_, err := service.LoginUser("email@gmail.com", "password")
	if !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUser_PasswordDoesNotMatch(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	fakerepo.GetUserByEmailError = apperrors.ErrInvalidCredentials
	password := "correctpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("error hashing password")
	}
	fakerepo.GetUserByEmailUser = models.User{
		ID:       1,
		Email:    "email@gmail.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	service := UserService{UserRepo: fakerepo}
	_, err = service.LoginUser("email@gmail.com", "wrongpassword")
	if !errors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUser_RepositoryError(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	fakerepo.GetUserByEmailError = errors.New("database error")
	service := UserService{UserRepo: fakerepo}
	_, err := service.LoginUser("email", "password")
	if !errors.Is(err, fakerepo.GetUserByEmailError) {
		t.Errorf("expected database error got %v", err)
	}
}

func TestLoginUser_Success(t *testing.T) {
	fakerepo := &FakeUserRepository{}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("error hashing password")
	}
	fakerepo.GetUserByEmailUser = models.User{
		ID:       1,
		Email:    "email@gmail.com",
		Password: string(hashedPassword),
		Role:     "user",
	}
	service := UserService{UserRepo: fakerepo}
	user, err := service.LoginUser("email@gmail.com", "password")
	if err != nil {
		t.Errorf("expected Success got %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Email != "email@gmail.com" {
		t.Errorf("expected email same but got %s", user.Email)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte("password"))
	if err != nil {
		t.Errorf("password was not hashed correctly")
	}
	if user.Role != "user" {
		t.Errorf("expected role same but got %s", user.Role)
	}
	if !fakerepo.GetUserByEmailCalled {
		t.Errorf("expected GetUserByEmailCalled to be true")
	}
}

package service
import(
	"errors"
	"golang.org/x/crypto/bcrypt"
	"task-manager-api/repository"
	"strings"
)
type UserService struct{
	UserRepo repository.UserRepo
}
var ErrEmptyEmail = errors.New("Email cannot be empty")
var ErrEmptyPassword = errors.New("Password cannot be empty")
var PasswordTooShort = errors.New("Password must be at least 6 characters")
var Password = errors.New("Password Hashed")
func (S *UserService) RegisterUser(email string, password string) error{
	email = strings.TrimSpace(email)
	if email == "" {
		return ErrEmptyEmail
	}
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 6 {
		return PasswordTooShort
	}
	hasedPassword,err := bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err != nil {
		return Password
	}
	return S.UserRepo.RegisterUser(email,string(hasedPassword))
}